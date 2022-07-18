package models

import (
	"time"

	"github.com/meteorae/meteorae-server/database"
)

type PersonType uint

const (
	PersonTypeUnknown PersonType = iota
	PersonTypeIndividual
	PersonTypeGroup
)

func (p PersonType) String() string {
	switch p {
	case PersonTypeIndividual:
		return "individual"
	case PersonTypeGroup:
		return "group"
	default:
		return "unknown"
	}
}

type Person struct {
	*MetadataModel
	Name      string
	SortName  string
	BirthDate time.Time
	DeathDate time.Time
	Type      PersonType
	Summary   string
	Thumb     string
	Art       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p Person) String() string {
	return p.Name
}

func NewPersonFromItemMetadata(m database.ItemMetadata) Person {
	person := Person{
		MetadataModel: &MetadataModel{
			ID: m.ID,
		},
		Name:      m.Title,
		BirthDate: m.ReleaseDate,
		DeathDate: m.EndDate,
		Type:      PersonTypeUnknown,
		Summary:   m.Summary,
		Thumb:     m.Thumb,
		Art:       m.Art,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.SortTitle != "" {
		person.SortName = m.SortTitle
	} else {
		person.SortName = m.Title
	}

	return person
}

func NewPersonSliceFromItemMetadata(m []database.ItemMetadata) []Person {
	people := make([]Person, len(m))

	for _, item := range m {
		people = append(people, NewPersonFromItemMetadata(item))
	}

	return people
}
