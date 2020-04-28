package migration
// TODO: Check connection before drop database
// TODO: lock database before run migration


import (
	"fmt"
	"time"
	"errors"
	"github.com/letoan96/psql_go_migration/adapter"
	"io/ioutil"
	"database/sql"
	"reflect"
	// "io"
)
var (
	directions = [2]string{"up", "down"}
)

type Migration struct {
	*adapter.DB
	Directory string

}

type NewMigration struct {
	Up string
	Down string
}

func Initialize(db *adapter.DB, migrationFolderPath string) Migration {
	migration := Migration {
		DB: db,
		Directory: migrationFolderPath,
	}
	return migration
}

func Generate(migration_name string, migrationFolderPath string) (*NewMigration, error) {
	t := time.Now()

	migrationFileName := fmt.Sprintf("%d%02d%02d%02d%02d%02d_%s",
        t.Year(), t.Month(), t.Day(),
        t.Hour(), t.Minute(), t.Second(), migration_name)  

    var migrationFileLocation string
   	var err error
	for _, direction := range directions {
    	migrationFileLocation = fmt.Sprintf("%s/%s.%s.sql", migrationFolderPath, migrationFileName, direction)
    	err = ioutil.WriteFile(migrationFileLocation, []byte(""), 0644)
	    if err != nil {
	        return nil,  err
	    }
	}

	newMigration := &NewMigration{
		Up: fmt.Sprintf("%s.up.sql", migrationFileName),
		Down : fmt.Sprintf("%s.down.sql", migrationFileName),
	}
	  
    return newMigration, nil
}


func (migration *Migration) Migrate() error {
	db := migration.DB

	// Will create table schema_migrations
	statement := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
		 	version varchar(20) NOT NULL,
		 	PRIMARY KEY (version)
		)
	`)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}

	// Get all migrate in migrate folder 
	migrateList, err := migration.ReadMigrateFolder()
	if err != nil {
		return err
	}
	
	// Run it
	err = migration.migrateUP(migrateList)
	// err = migration.migrateDown(migrateList, 1)
	if err != nil {
		return err 
	}
	
	return nil
}

func (migration *Migration) RollBack(step int) error {
	// Get all migrate in migrate folder 
	migrateList, err := migration.ReadMigrateFolder()
	if err != nil {
		return err
	}

	// Run it
	fmt.Printf("*** Rolling back last %v  migration ***\n", step)
	err = migration.migrateDown(migrateList, step)
	if err != nil {
		return err 
	}
	
	return nil
}

// run the migrate
func (migration *Migration) run(migrate *Migrate) error {
	db := migration.DB
	dat, _ := ioutil.ReadFile(migrate.Path)
	// fmt.Printf("\nData: %s", dat)
	statement := string(dat)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}

	fmt.Printf("== %s: done ======================================\n", migrate.Name)
	return nil
	// then write version to schema_migration
}

// Migrate all the way up
func (migration *Migration) migrateUP(migrateList *MigrateList) error {
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
    	if migrate.Direction == "up" {
    		err := migration.run(migrate)
    		if err != nil {
    			return err
    		}
    		migration.appendMigrateVersion(migrate.Version)
    	}
	}

    return nil
}

// Migrate down from current version
func (migration *Migration) migrateDown(migrateList *MigrateList, step int) error {
	currentVersion := migration.getCurrentVersion()
	downList := MigrateList{} // migrations are going to rollback
	for i, migrate := range *migrateList {
		j := i + 1
    	if migrate.Version == currentVersion {
        	downList = (*migrateList)[:j]
        	break
    	}
    }
    // iterate in reverse
    for i := len(downList)-1; i >= 0; i-- {
    	if step <= 0 {
    		return nil
    	}
    	migrate := downList[i]
   		// fmt.Println(s[i])
   		if migrate.Direction == "down" {
    		err := migration.run(migrate)
    		if err != nil {
    			return err
    		}
    		migration.deleteMigrateVersion(migrate.Version)
    		step--
    	}

	}
    return nil
}

// The function name has spoken for itself 
func (migration *Migration) CreatDatabaseIfNotExists() error {
	db := migration.DB
	if migration.doesDatabaseExist() == true {
		return errors.New(fmt.Sprintf("Database '%s' already exists.", db.Database))
	} else {
		statement := fmt.Sprintf("CREATE DATABASE %s;", db.Database)
    	_, err := db.Exec(statement)
    	if err != nil {
			return err
		}
	}

	return nil
} 

// The function name has spoken for itself 
func (migration *Migration) DropDatabase() error { 
	db := migration.DB
	if migration.doesDatabaseExist() == true {
		statement := fmt.Sprintf("DROP DATABASE %s;", db.Database)
    	_, err := db.Exec(statement)
    	if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("Database '%s' does not exists", db.Database))
	}
	return nil
}

func (migration *Migration) doesDatabaseExist() bool {
	db := migration.DB
	statement := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", db.Database)
	row := db.QueryRow(statement)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		panic(err)
	}
	return exists
}

// Get current migrate version in  schema_migrations
func (migration *Migration) getCurrentVersion() string {
	db := migration.DB
	statement := fmt.Sprintf(`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1 `)
	row := db.QueryRow(statement)
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

func (migration *Migration) getSchemaMigrations() []string{
	list := []string{}
	db := migration.DB
	statement := fmt.Sprintf(`SELECT version FROM schema_migrations ORDER BY version DESC`)
	rows, err := db.Query(statement)
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

		list = append(list, version)
	}
	return list
}

func (migration *Migration) appendMigrateVersion(version string) {
	db := migration.DB
	statement := fmt.Sprintf("INSERT INTO schema_migrations (version) VALUES ('%s');", version)
	_, err := db.Exec(statement)
	if err != nil {
		panic(err)
	}
}

func (migration *Migration) deleteMigrateVersion(version string) {
	db := migration.DB
	statement := fmt.Sprintf("DELETE FROM schema_migrations WHERE version='%s';", version)
	_, err := db.Exec(statement)
	if err != nil {
		panic(err)
	}
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