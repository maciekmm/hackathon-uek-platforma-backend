package timetable

import (
	"errors"
	"strconv"
	"time"

	"regexp"

	"strings"

	"github.com/PuerkitoBio/goquery"
)

const delay = 500 * time.Millisecond

var idRegexp = regexp.MustCompile("id=(\\d+)")

func ScrapeGroupIDs() (associations map[string]map[string]int, err error) {
	associations = map[string]map[string]int{}
	doc, err := goquery.NewDocument(basePath)
	if err != nil {
		return nil, err
	}
	doc.Find(".kategorie > a").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		categoryURL, ok := sel.Attr("href")
		if !ok || !strings.Contains(categoryURL, "typ=G") {
			return true
		}
		categoryURL = basePath + categoryURL
		groupName := sel.Text()
		if len(groupName) == 0 {
			return true
		}
		associations[groupName] = map[string]int{}
		subDoc, iErr := goquery.NewDocument(categoryURL)
		if iErr != nil {
			err = iErr
			return false
		}
		time.Sleep(delay)
		subDoc.Find(".kolumny a").EachWithBreak(func(i int, sel *goquery.Selection) bool {
			name := sel.Text()
			if rawID, ok := sel.Attr("href"); ok {
				if parsed, iErr := strconv.Atoi(idRegexp.FindStringSubmatch(rawID)[1]); iErr == nil {
					associations[groupName][name] = parsed
				} else {
					err = iErr
					return false
				}
				return true
			}
			err = errors.New("timetable format might have changed")
			return false
		})
		return err == nil
	})
	return associations, nil
}
