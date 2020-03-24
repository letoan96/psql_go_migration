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
    	err = ioutil.WriteFile(migrationFileLocation, []byte("BEGIN\n-------PLACE YOUR STATEMENT INSIDE BEGIN-END BLOCK-------\n\nEND"), 0644)
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

	migrateList, err := migration.ReadMigrateFolder()
	if err != nil {
		return err
	}

	err = migrateList.migrateUP()
	if err != nil {
		return err
	}
	
	return nil
}

// run the migrate
func (migrate Migrate) run() {
	fmt.Printf("%+v\n", migrate)
	// then write version to schema_migration
}

// Migrate all the way up
func (migrateList *MigrateList) migrateUP() error {
	currentVersion := migration.getCurrentVersion()
	upList := MigrateList{} // migrations are going to migrate
	for i, migrate := range *migrateList {
    	if migrate.Version == currentVersion && migrate.Direction == "up" {
    		j := i + 1
        	upList = (*migrateList)[j:]
        	break
    	}
    }



    for _, migrate := range upList {
    	if migrate.Direction == "up" {
    		migrate.run()  
    	}
	}

    return nil
}

// Migrate down from current version
func (migrateList *MigrateList) migrateDown(step int) error {
	// currentVersion := migration.getCurrentVersion()
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

// The function name has spoken for itself 
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
