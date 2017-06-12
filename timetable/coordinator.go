package timetable

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"os"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

const minimumTimetableLongevity = 2 * time.Hour
const pathPrefix = "/var/lib/uek/"

type Coordinator struct {
	Logger       *log.Logger
	Database     *gorm.DB
	ticker       *time.Ticker
	stop         chan interface{}
	cache        map[string]*Timetable
	associations []byte
}

func NewCoordinator(interval time.Duration, database *gorm.DB, logger *log.Logger) *Coordinator {
	return &Coordinator{
		Logger:   logger,
		Database: database,
		ticker:   time.NewTicker(interval),
		stop:     make(chan interface{}),
		cache:    make(map[string]*Timetable),
	}
}

func (c *Coordinator) Register(router *mux.Router) error {
	router.HandleFunc("/groups", c.GetAssociations).Methods(http.MethodGet)
	return nil
}

func (c *Coordinator) GetAssociations(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write(c.associations)
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

func (c *Coordinator) checkUpdates() {
	res := []uint{}
	c.Database.Raw("SELECT DISTINCT \"group\" FROM \"users\"").Scan(&res)
	for _, group := range res {
		c.Logger.Printf("checking plan updates for %d-%d\n", group, 3)
		old, ok := c.cache[fmt.Sprintf("%d-%d", group, 3)]
		new, err := c.Load(group, 3, true)
		if err != nil {
			c.Logger.Printf("could not fetch and parse timetable for group: %d, period: %d\n", group, 3)
			continue
		}
		if ok {
			diff := old.Diff(new)
			fmt.Println(diff)
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Coordinator) Stop() {
	c.ticker.Stop()
}

func (c *Coordinator) Load(group uint, period uint, force bool) (*Timetable, error) {
	if tt, ok := c.cache[fmt.Sprintf("%d-%d", group, period)]; !force && ok {
		return tt, nil
	}
	file, err := os.Open(fmt.Sprintf("%s%d-%d.json", pathPrefix+"timetables/", group, period))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// decode
	tt := Timetable{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tt); err != nil {
		return nil, err
	}
	c.cache[fmt.Sprintf("%d-%d", group, period)] = &tt
	return &tt, nil
}

func (c *Coordinator) Save(tt *Timetable, period uint) error {
	file, err := os.OpenFile(fmt.Sprintf("%s%d-%d.json", pathPrefix+"timetables/", tt.GroupID, period), os.O_CREATE|os.O_RDWR, 0755)
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
	return nil
}
