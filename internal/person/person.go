package person

import "time"

type Person struct {
	name        string
	surname     string
	patronimic  string
	group       string
	wallet      string
	dateOfBirth time.Time
}

func NewPerson(name, surname, patronimic, group, wallet string, dateOfBirth time.Time) *Person {
	return &Person{
		name:        name,
		surname:     surname,
		patronimic:  patronimic,
		group:       group,
		wallet:      wallet,
		dateOfBirth: dateOfBirth,
	}
}
