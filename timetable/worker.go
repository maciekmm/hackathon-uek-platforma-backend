package timetable

import (
	"fmt"
	"log"
	"time"

	"os"

	"encoding/json"

	"github.com/jinzhu/gorm"
)

const minimumTimetableLongevity = 2 * time.Hour
const pathPrefix = "/var/lib/uek/timetables/"

type Worker struct {
	Logger   *log.Logger
	Database *gorm.DB
	ticker   *time.Ticker
	stop     chan interface{}
	cache    map[string]*Timetable
}

func NewWorker(interval time.Duration, database *gorm.DB, logger *log.Logger) *Worker {
	return &Worker{
		Logger:   logger,
		Database: database,
		ticker:   time.NewTicker(interval),
		stop:     make(chan interface{}),
		cache:    make(map[string]*Timetable),
	}
}

func (w *Worker) Start() error {
	err := os.MkdirAll(pathPrefix, 0755)
	if err != nil {
		return err
	}
	for {
		select {
		case t, ok := <-w.ticker.C:
			if !ok {
				w.Logger.Println("stopping timetable worker")
				return nil
			}
			w.Logger.Printf("checking for timetable updates: %s", t.Format(timeFormat))
			res := []uint{}
			w.Database.Raw("SELECT DISTINCT \"group\" FROM \"users\"").Scan(&res)
			for _, group := range res {
				old, ok := w.cache[fmt.Sprintf("%d-%d", group, 3)]
				new, err := w.Load(group, 3, true)
				if err != nil {
					w.Logger.Printf("could not fetch and parse timetable for group: %d, period: %d", group, 3)
					continue
				}
				if ok {
					diff := old.Diff(new)
					fmt.Println(diff)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func (w *Worker) Stop() {
	w.ticker.Stop()
}

func (w *Worker) Load(group uint, period uint, force bool) (*Timetable, error) {
	if tt, ok := w.cache[fmt.Sprintf("%d-%d", group, period)]; !force && ok {
		return tt, nil
	}
	file, err := os.Open(fmt.Sprintf("%s%d-%d.json", pathPrefix, group, period))
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
	w.cache[fmt.Sprintf("%d-%d", group, period)] = &tt
	return &tt, nil
}

func (w *Worker) Save(tt *Timetable, period uint) error {
	file, err := os.OpenFile(fmt.Sprintf("%s%d-%d.json", pathPrefix, tt.GroupID, period), os.O_CREATE|os.O_RDWR, 0755)
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
