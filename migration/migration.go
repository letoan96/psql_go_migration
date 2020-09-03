package migration

// TODO: lock database before run migration

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/letoan96/psql_go_migration/task"
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
	defer timeTrack(time.Now(), "Migration")
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
	migrateList := migration.readMigrateFolder()
	// Run it
	migration.migrateUP(migrateList)
}

func (migration *Migration) RollBack(step int) {
	if step <= 0 {
		return
	}

	migrateList := migration.readMigrateFolder() // Get all migrate in migrate folder
	fmt.Printf("*** Rolling back last %v  migration ***\n", step)
	migration.migrateDown(migrateList, step)
}

// run the migrate
func (migration *Migration) runUp(migrate *MigrateFile) {
	content, err := ioutil.ReadFile(migrate.Path)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(content))
	var statementBuffer bytes.Buffer
	var beforeTask, afterTask []string

	for scanner.Scan() {
		s := scanner.Text()
		if strings.Contains(s, "before_migrate:") {
			a := strings.Split(s, "before_migrate:")
			beforeTask = unmarshal(a[1])
		} else if strings.Contains(s, "after_migrate:") {
			a := strings.Split(s, "after_migrate:")
			afterTask = unmarshal(a[1])
		} else {
			statementBuffer.WriteString(s)
			statementBuffer.WriteString(" ")
		}
	}

	task.RunTask(beforeTask)

	trx, err := migration.DB.Begin()
	if err != nil {
		panic(err)
	}
	defer trx.Rollback()

	statement := statementBuffer.String()
	fmt.Println("======================", statement)
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

	task.RunTask(afterTask)
	fmt.Printf("== %s: done ======================================\n", migrate.Name)
}

func (migration *Migration) migrateUP(migrateList *MigrateList) {
	migratedMap := migration.getSchemaMigrations()
	upList := MigrateList{} // migrations which are going to migrate
	if len(migratedMap) == 0 {
		upList = *migrateList
	} else {
		for _, migrate := range *migrateList {
			_, ok := migratedMap[migrate.Version]
			if !ok && migrate.Direction == "up" {
				upList = append(upList, migrate)
			}
		}
	}

	for _, migrate := range upList {
		if migrate.Direction == "up" {
			migration.runUp(migrate)
		}
	}
}

// Migrate down from current version
func (migration *Migration) migrateDown(migrateList *MigrateList, step int) {
	downMap := migration.getPreviousVersion(step)
	downList := MigrateList{} // migrations are going to rollback

	for _, migrate := range *migrateList {
		_, ok := downMap[migrate.Version]
		if ok && migrate.Direction == "down" {
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

func (migration *Migration) getSchemaMigrations() map[string]int {
	migationMap := map[string]int{}
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

	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			panic(err)
		}

		migationMap[version] = 1
	}

	return migationMap
}

func (migration *Migration) getPreviousVersion(step int) map[string]int {
	migationMap := map[string]int{}
	rows, err := migration.DB.Query(`
		SELECT 
			version
		FROM
			schema_migrations
		ORDER BY
			version DESC
		LIMIT
			$1`, step)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			panic(err)
		}

		migationMap[version] = 1
	}

	return migationMap
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

func unmarshal(s string) []string {
	var arr []string
	err := json.Unmarshal([]byte(s), &arr)
	if err != nil {
		panic(errors.New("Can't not unmarshal task list in migration"))
	}
	return arr
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
