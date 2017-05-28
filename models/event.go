package models

import (
	"time"
)

type Event struct {
	ID              uint64    `json:"id,omitempty" db:"id"`
	Owner           uint64    `json:"owner,omitempty" db:"owner"`
	Name            string    `json:"name,omitempty" db:"name"`
	Description     string    `json:"description,omitempty" db:"description"`
	DateAdded       time.Time `json:"date_added,omitempty" db:"date_added"`
	DateTakingPlace time.Time `json:"date_taking_place,omitempty" db:"date_taking_place"`
	DateUpdated     time.Time `json:"date_updated,omitempty" db:"date_updated"`
}
