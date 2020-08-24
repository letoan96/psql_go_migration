package migration

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/fatih/color"
)

type MigrateFile struct {
	Name      string
	Version   string
	Direction string
	Path      string
}

type MigrateList []*MigrateFile

var (
	Regex = regexp.MustCompile(`^([0-9]+)_(.*)\.(` + `down` + `|` + `up` + `)\.(.*)$`)
)

func (migration *Migration) readMigrateFolder() *MigrateList {
	migrateList := MigrateList{}

	files, err := ioutil.ReadDir(migration.Directory)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			migrateFile, err := Parse(file.Name())
			if err != nil {
				color.Red(fmt.Sprintf(`Can not read file: '%s'`, file.Name()))
				continue // ignore files that can't be parsed
			}
			migrateFile.Path = fmt.Sprintf("%s/%s", migration.Directory, file.Name())
			migrateList = append(migrateList, migrateFile)
		}
	}

	return &migrateList
}

func Parse(raw string) (*MigrateFile, error) {
	m := Regex.FindStringSubmatch(raw)
	if len(m) == 5 {
		return &MigrateFile{
			Version:   string(m[1]),
			Name:      m[2],
			Direction: m[3],
		}, nil
	}

	return nil, fmt.Errorf("Migrate file no match")
}
