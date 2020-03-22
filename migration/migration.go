package migration

import (
	"fmt"
	"errors"
	"github.com/letoan96/psql_go_migration/adapter"

)

type Migration struct {
	*adapter.DB
	Directory string
}

func Initialize(db *adapter.DB, migrationFolderPath string) Migration {
	migration := Migration {
		DB: db,
		Directory: migrationFolderPath,
	}
	return migration
}

func (migration *Migration) Migrate() {
	fmt.Println(`It works`)
}

func (migration *Migration) CreatDatabaseIfNotExists() error {
	statement := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", migration.DB.Database)
	row := migration.DB.QueryRow(statement)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return err
	}

	if exists == true {
		return errors.New(fmt.Sprintf("Database '%s' already exists.", migration.DB.Database))
	} else {
		statement = fmt.Sprintf("CREATE DATABASE %s;", migration.DB.Database)
    	_, err = migration.DB.Exec(statement)
    	if err != nil {
			return err
		}
	}
	return nil
	// if exists == false {
	//     statement = `CREATE DATABASE yourDBName;`
	//     _, err = db.Exec(statement)
	//     check(err)
	// }
} 