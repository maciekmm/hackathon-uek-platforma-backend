package timetable

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"os"

	"encoding/json"

	"sync"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/maciekmm/uek-bruschetta/middleware"
	"github.com/maciekmm/uek-bruschetta/models"
	"github.com/maciekmm/uek-bruschetta/utils"
)

const minimumTimetableLongevity = 2 * time.Hour
const pathPrefix = "/var/lib/uek/"

var (
	ErrTimetableUnknown = errors.New("unknown error occured")
)

type Coordinator struct {
	Logger       *log.Logger
	Database     *gorm.DB
	ticker       *time.Ticker
	stop         chan interface{}
	cache        map[string]*Timetable
	mapMutex     *sync.RWMutex
	associations []byte
}

func NewCoordinator(interval time.Duration, database *gorm.DB, logger *log.Logger) *Coordinator {
	return &Coordinator{
		Logger:   logger,
		Database: database,
		ticker:   time.NewTicker(interval),
		stop:     make(chan interface{}),
		mapMutex: &sync.RWMutex{},
		cache:    make(map[string]*Timetable),
	}
}

func (c *Coordinator) Register(router *mux.Router) error {
	router.HandleFunc("/groups/", c.HandleGetAssociations).Methods(http.MethodGet)
	router.HandleFunc("/{group:[0-9]+}/{period:[0-9]+}/", c.HandleGetTimetable).Methods(http.MethodGet)
	router.Handle("/", middleware.RequiresAuth(models.RoleUser, http.HandlerFunc(c.HandleGetTimetable))).Methods(http.MethodGet)
	return nil
}

func (c *Coordinator) HandleGetAssociations(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write(c.associations)
}

func (c *Coordinator) HandleGetTimetable(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var group *uint
	period := uint(3)

	if id, err := strconv.Atoi(vars["group"]); err != nil {
		if user, ok := r.Context().Value(middleware.ContextUserKey).(*models.User); ok {
			group = user.Group
		}
	} else {
		per, _ := strconv.Atoi(vars["period"])
		period = uint(per)
		temp := uint(id)
		group = &temp
	}

	if group == nil {
		utils.NewErrorResponse(errors.New("no group id specified for this user")).Write(http.StatusBadRequest, rw)
		return
	}

	tt, _, err := c.Load(*group, period, false)
	if err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{ErrTimetableUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
	enc := json.NewEncoder(rw)
	if err := enc.Encode(tt); err != nil {
		(&utils.ErrorResponse{
			Errors:      []string{ErrTimetableUnknown.Error()},
			DebugErrors: []string{err.Error()},
		}).Write(http.StatusInternalServerError, rw)
		return
	}
}

func (c *Coordinator) Start() error {
	err := os.MkdirAll(pathPrefix+"groups/", 0755)
	if err != nil {
		return err
	}
	//Fetch group associations if not present
	if f, err := os.Open(pathPrefix + "groups/group-assoc.json"); err != nil && os.IsNotExist(err) {
		c.Logger.Println("fetching group ids for reference")
		f, err := os.Create(pathPrefix + "groups/group-assoc.json")
		if err != nil {
			return fmt.Errorf("could not create group associations file: %s", err.Error())
		}
		defer f.Close()
		groups, err := ScrapeGroupIDs()
		if err != nil {
			return fmt.Errorf("could not scrape group assocaitions: %s", err.Error())
		}
		byt, err := json.Marshal(&groups)
		if err != nil {
			return fmt.Errorf("could not encode group associations: %s", err.Error())
		}
		if _, err := f.Write(byt); err != nil {
			return fmt.Errorf("could not write associations to file: %s", err.Error())
		}
		c.associations = byt
	} else if err == nil {
		defer f.Close()
		byt, err := ioutil.ReadFile(pathPrefix + "groups/group-assoc.json")
		if err != nil {
			return fmt.Errorf("could not read group associations: %s", err.Error())
		}
		c.associations = byt
	}

	c.Logger.Println("Starting update check routine")
	//start the routine
	err = os.MkdirAll(pathPrefix+"timetables/", 0755)
	if err != nil {
		return err
	}
	return nil
	c.checkUpdates()
	for {
		select {
		case _, ok := <-c.ticker.C:
			if !ok {
				c.Logger.Println("stopping timetable worker")
				return nil
			}
			c.checkUpdates()
		}
	}
}

func (c *Coordinator) checkUpdates() error {
	res := []uint{}
	rows, err := c.Database.Raw("SELECT DISTINCT \"group\" FROM \"users\"").Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var i uint
		if err = rows.Scan(&i); err != nil {
			return err
		}
		res = append(res, i)
	}

	for _, group := range res {
		c.Logger.Printf("checking plan updates for %d-%d\n", group, 3)
		old, cached, err := c.Load(group, 3, false)
		if err != nil {
			c.Logger.Printf("could not fetch and parse timetable for group: %d, period: %d, err: %s\n", group, 3, err.Error())
			continue
		}
		if !cached {
			continue
		}
		new, _, err := c.Load(group, 3, true)
		if err != nil {
			c.Logger.Printf("could not fetch and parse timetable for group: %d, period: %d, err: %s\n", group, 3, err.Error())
			continue
		}
		diff := old.Diff(new)
		if len(diff) > 0 {
			fmt.Println("diff detected")
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (c *Coordinator) Stop() {
	c.ticker.Stop()
}

func (c *Coordinator) Load(group uint, period uint, force bool) (*Timetable, bool, error) {
	// check if timetable is in cache already
	c.mapMutex.RLock()
	if tt, ok := c.cache[fmt.Sprintf("%d-%d", group, period)]; !force && ok {
		c.mapMutex.RUnlock()
		return tt, true, nil
	}
	c.mapMutex.RUnlock()

	var tt *Timetable

	// if forcing update fetch the new timetable
	if force {
		parsed, err := TimetableFromId(group, period)
		if err != nil {
			return nil, false, err
		}
		tt = parsed
		if err := c.Save(parsed, period); err != nil {
			return nil, false, err
		}
	}

	if tt == nil {
		// open the file we might have saved to, TODO: fix this
		file, err := os.Open(fmt.Sprintf("%s%s%d-%d.json", pathPrefix, "timetables/", group, period))
		if err != nil {
			return c.Load(group, period, true)
		}
		defer file.Close()
		base := &Timetable{}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(base); err != nil {
			return nil, false, err
		}
		tt = base
	}

	// decode
	c.mapMutex.Lock()
	c.cache[fmt.Sprintf("%d-%d", group, period)] = tt
	c.mapMutex.Unlock()
	return tt, !force, nil
}

func (c *Coordinator) Save(tt *Timetable, period uint) error {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	file, err := os.OpenFile(fmt.Sprintf("%s%s%d-%d.json", pathPrefix, "timetables/", tt.GroupID, period), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	byt, err := json.Marshal(tt)
	if err != nil {
		return err
	}
	_, err = file.Write(byt)
	if err != nil {
		return err
	}
	return file.Close()
}
