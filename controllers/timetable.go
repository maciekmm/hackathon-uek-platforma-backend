package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"io/ioutil"

	"github.com/gorilla/mux"
	"github.com/maciekmm/uek-bruschetta/timetable"
)

const groupAssociations = "./data/group-associations.json"

type Timetable struct {
	Logger       *log.Logger
	associations []byte
}

func (t *Timetable) Register(router *mux.Router) error {
	// timetables TODO: move this somewhere else
	if f, err := os.Open(groupAssociations); err != nil && os.IsNotExist(err) {
		t.Logger.Printf("fetching group ids for reference")
		f, err := os.Create(groupAssociations)
		if err != nil {
			return fmt.Errorf("could not create group associations file: %s", err.Error())
		}
		defer f.Close()
		groups, err := timetable.ScrapeGroupIDs()
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
		t.associations = byt
	} else if err == nil {
		defer f.Close()
		byt, err := ioutil.ReadFile(groupAssociations)
		if err != nil {
			return fmt.Errorf("could not read group associations: %s", err.Error())
		}
		t.associations = byt
	}

	router.HandleFunc("/groups", t.GetAssociations).Methods(http.MethodGet)
	return nil
}

func (t *Timetable) GetAssociations(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write(t.associations)
}
