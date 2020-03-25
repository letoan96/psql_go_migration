package migration

import (
	"fmt"
	"regexp"
	"io/ioutil"
)
type Migrate struct {
	Name string
	Version string
	Direction string
	Path string
}

type MigrateList []*Migrate

func (migration *Migration) ReadMigrateFolder() (*MigrateList, error) {
	migrateList := MigrateList{}
	
	files, err := ioutil.ReadDir(migration.Directory) // scan directory
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			m, err := Parse(file.Name())
			if err != nil {
				continue // ignore files that can't be parsed
			}
			m.Path = fmt.Sprintf("%s/%s", migration.Directory, file.Name())
			migrateList = append(migrateList, m)
		}
	}

	return &migrateList, nil
}

// Parse returns Migration for matching Regex pattern.
var (
	ErrParse = fmt.Errorf("Migrate file no match")
	Regex = regexp.MustCompile(`^([0-9]+)_(.*)\.(` + `down` + `|` + `up` + `)\.(.*)$`)
)

func Parse(raw string) (*Migrate, error) {
	m := Regex.FindStringSubmatch(raw)
	if len(m) == 5 {
		versionstring := string(m[1])
		return &Migrate{
			Version:    string(versionstring),
			Name: m[2],
			Direction:  m[3],
		}, nil
	}
	return nil, ErrParse
}

