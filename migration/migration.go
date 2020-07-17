package migration

// TODO: Check connection before drop database
// TODO: lock database before run migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/fatih/color"
	// "io"
)

var (
	directions = [2]string{"up", "down"}
)

type Migration struct {
	*sql.DB
	Directory string
}

type NewMigration struct {
	Up   string
	Down string
}

func Initialize(db *sql.DB, folderPath string) *Migration {
	migration := &Migration{
		DB:        db,
		Directory: folderPath,
	}
	return migration
}

func Generate(migrationName, folderPath string) {
	t := time.Now()

	fileName := fmt.Sprintf("%d%02d%02d%02d%02d%02d_%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), migrationName)

	for _, direction := range directions {
		fileLocation := fmt.Sprintf("%s/%s.%s.sql", folderPath, fileName, direction)
		err := ioutil.WriteFile(fileLocation, []byte(""), 0644)
		if err != nil {
			panic(err)
		}

	}

	color.Green("---> Created %s", fmt.Sprintf("%s.up.sql", fileName))
	color.Green("---> Created %s", fmt.Sprintf("%s.down.sql", fileName))

}

func (migration *Migration) Migrate() {
	// Will create table schema_migrations
	_, err := migration.DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
		 	version varchar(20) NOT NULL,
		 	PRIMARY KEY (version)
		)`)
	if err != nil {
		panic(err)
	}

	// Get all migrate in migrate folder
	migrateList := migration.ReadMigrateFolder()
	// Run it
	migration.migrateUP(migrateList)
}

func (migration *Migration) RollBack(step int) {
	if step <= 0 {
		return
	}

	migrateList := migration.ReadMigrateFolder() // Get all migrate in migrate folder
	fmt.Printf("*** Rolling back last %v  migration ***\n", step)
	migration.migrateDown(migrateList, step)
}

// run the migrate
func (migration *Migration) runUp(migrate *MigrateFile) {
	statement, err := ioutil.ReadFile(migrate.Path)
	if err != nil {
		panic(err)
	}

	trx, err := migration.DB.Begin()
	if err != nil {
		panic(err)
	}
	defer trx.Rollback()

	_, err = trx.Exec(string(statement))
	if err != nil {
		panic(err)
	}

	err = appendMigrateVersion(trx, migrate.Version)
	if err != nil {
		panic(err)
	}

	err = trx.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Printf("== %s: done ======================================\n", migrate.Name)
}

// Migrate all the way up
func (migration *Migration) migrateUP(migrateList *MigrateList) {
	migratedList := migration.getSchemaMigrations()
	upList := MigrateList{} // migrations which are going to migrate

	if len(migratedList) == 0 {
		upList = *migrateList
	} else {
		for _, migrate := range *migrateList {
			if !itemExists(migratedList, migrate.Version) && migrate.Direction == "up" {
				upList = append(upList, migrate)
			}
		}
	}

	for _, migrate := range upList {
		migration.runUp(migrate)
	}
}

// Migrate down from current version
func (migration *Migration) migrateDown(migrateList *MigrateList, step int) {
	downVersions := migration.getPreviousVersion(step)
	downList := MigrateList{} // migrations are going to rollback

	for _, migrate := range *migrateList {
		if itemExists(downVersions, migrate.Version) && migrate.Direction == "down" {
			downList = append(downList, migrate)
		}
	}
	// iterate in reverse
	for _, migrate := range downList {
		migration.runDown(migrate)
	}
}

func (migration *Migration) runDown(migrate *MigrateFile) {
	statement, err := ioutil.ReadFile(migrate.Path)
	if err != nil {
		panic(err)
	}

	trx, err := migration.DB.Begin()
	if err != nil {
		panic(err)
	}
	defer trx.Rollback()

	_, err = trx.Exec(string(statement))
	if err != nil {
		panic(err)
	}

	err = deleteMigrateVersion(trx, migrate.Version)
	if err != nil {
		panic(err)
	}

	err = trx.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Printf("== %s: done ======================================\n", migrate.Name)
}

// Get current migrate version in  schema_migrations
func (migration *Migration) getCurrentVersion() string {
	statement := fmt.Sprintf(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1 `)
	row := migration.DB.QueryRow(statement)
	var version string
	err := row.Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "-1"
		} else {
			panic(err)
		}
	}
	return version
}

func (migration *Migration) getSchemaMigrations() []string {
	rows, err := migration.DB.Query(`
		SELECT 
			version
		FROM
			schema_migrations
		ORDER BY
			version DESC`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	list := []string{}
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			panic(err)
		}

		list = append(list, version)
	}
	return list
}

func (migration *Migration) getPreviousVersion(step int) []string {
	rows, err := migration.DB.Query(`
		SELECT 
			version
		FROM
			schema_migrations
		ORDER BY
			version DESC
		LIMIT
			$1
		OFFSET 1`, step)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	list := []string{}
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			panic(err)
		}

		list = append(list, version)
	}
	return list
}

func appendMigrateVersion(trx *sql.Tx, version string) error {
	_, err := trx.Exec(`
		INSERT INTO 
			schema_migrations (version) 
		VALUES ($1)`, version)
	if err != nil {
		return err
	}

	return nil
}

func deleteMigrateVersion(trx *sql.Tx, version string) error {
	_, err := trx.Exec(`
		DELETE FROM
			schema_migrations
		WHERE
			version = $1`, version)
	if err != nil {
		return err
	}

	return nil
}

func itemExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)
	if arr.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}
	return false
}
