package timetable

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const basePath = "http://planzajec.uek.krakow.pl/"

var hourExpression = regexp.MustCompile("[\\d]{2}:[\\d]{2}")

const timeFormat = "2006-01-02 15:04"

type ParsingError struct {
	Errors []error
}

func (pe *ParsingError) Error() string {
	str := []string{}
	for _, err := range pe.Errors {
		str = append(str, err.Error())
	}
	return strings.Join(str, ", ")
}

type Timetable struct {
	GroupID uint     `json:"group_id"`
	Group   string   `json:"group"`
	Classes []*Class `json:"classes"`
}

type Class struct {
	Start   time.Time `json:"start,omitempty"`
	End     time.Time `json:"end,omitempty"`
	Class   string    `json:"class,omitempty"`
	Type    string    `json:"type,omitempty"`
	Teacher string    `json:"teacher,omitempty"`
	Room    string    `json:"room,omitempty"`
	Note    string    `json:"note,omitempty"`
	Urgent  bool      `json:"urgent,omitempty"`
}

func (c *Class) Valid() bool {
	return !c.Start.IsZero() && !c.End.IsZero() && len(c.Class) > 0
}

func (c *Class) Equal(new *Class) bool {
	return c.Start.Equal(new.Start) && c.End.Equal(new.End) && c.Class == new.Class && c.Type == new.Type && c.Teacher == new.Teacher && c.Room == new.Room && c.Note == new.Note
}

func TimetableFromId(id uint, period uint) (*Timetable, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%sindex.php?typ=G&id=%d&okres=%d", basePath, id, period))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}
	defer resp.Body.Close()
	return ParseTimetable(resp.Body, id)
}

func ParseTimetable(reader io.Reader, id uint) (*Timetable, error) {
	timetable := &Timetable{
		Classes: []*Class{},
		GroupID: id,
	}
	errs := []error{}
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	timetable.Group = doc.Find(".grupa").Text()
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		// skip the header row
		if i == 0 {
			return
		}
		class := Class{}
		if s.HasClass("czerwony") {
			class.Urgent = true
		}
		termin := ""
		s.Find("td").EachWithBreak(func(j int, sel *goquery.Selection) bool {
			if sel.HasClass("uwagi") && len(timetable.Classes) > 0 {
				timetable.Classes[len(timetable.Classes)-1].Note = sel.Text()
				return false
			}
			if sel.AttrOr("colspan", "") != "" {
				return false
			}

			switch j {
			//termin
			case 0:
				termin = sel.Text()
			//godzina
			case 1:
				hours := hourExpression.FindAllString(sel.Text(), -1)
				if len(hours) != 2 {
					errs = append(errs, fmt.Errorf("invalid hours field: %s", sel.Text()))
					return false
				}
				parsed, err := time.Parse(timeFormat, termin+" "+hours[0])
				if err != nil {
					errs = append(errs, fmt.Errorf("could not parse time: %s", sel.Text()))
					return false
				}
				class.Start = parsed
				parsed, err = time.Parse(timeFormat, termin+" "+hours[1])
				if err != nil {
					errs = append(errs, fmt.Errorf("could not parse time: %s", sel.Text()))
					return false
				}
				class.End = parsed
			//przedmiot
			case 2:
				class.Class = sel.Text()
			case 3:
				class.Type = sel.Text()
			case 4:
				class.Teacher = sel.Text()
			case 5:
				class.Room = sel.Text()
			}

			return true
		})
		if class.Valid() {
			timetable.Classes = append(timetable.Classes, &class)
		}
	})
	return timetable, &ParsingError{Errors: errs}
}

type ClassDiff struct {
	Old *Class `json:"old"`
	New *Class `json:"new"`
}

func (old *Timetable) Diff(new *Timetable) []ClassDiff {
	cd := []ClassDiff{}
	for i, newClass := range new.Classes {
		//direct match (?)
		oldClass := old.Classes[i]
		if oldClass.Equal(newClass) {
			continue
		}
		//support shifting
		if !oldClass.Start.Equal(newClass.Start) {
			oldClass = nil
			for _, class := range old.Classes {
				if class.Start.Equal(newClass.Start) {
					oldClass = class
					break
				}
			}
		}
		if oldClass != nil && !oldClass.Equal(newClass) {
			cd = append(cd, ClassDiff{
				Old: oldClass,
				New: newClass,
			})
		} else if oldClass == nil {
			cd = append(cd, ClassDiff{
				Old: nil,
				New: newClass,
			})
		}
	}
	return cd
}
