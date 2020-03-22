package main

import (
	"fmt"
	"github.com/letoan96/psql_go_migration/adapter"
	"github.com/letoan96/psql_go_migration/migration"
	"github.com/fatih/color"
)

// type Database struct {
// 	instance *adapter.DB
// }


func main() {
	config := map[string]string{
		"type": "postgres",
		"database":    "p2play2",
		"username":    "dever",
		"password":    "dever",
		"host":        "127.0.0.1",
		"port":		   "5433",
		"maxIdleConnection": "80",
		"maxOpenConnection": "40",
	}
	p2playAdapter := adapter.Initialize(config)
	// db, err := p2playAdapter.ConnectToDatabase()
	// if err != nil {
	// 	return
	// }

	db, err := p2playAdapter.ConnectToPostgres()
	if err != nil {
		panic(err)
	}

	p2playMigration := migration.Initialize(db, "/db/migrations")
	err = p2playMigration.CreatDatabaseIfNotExists()
	if err != nil {
		red := color.New(color.FgRed).PrintfFunc()
		red("%s\n", err)
		panic(err)
	}
	color.Green("Database created")


	db.Close()
	fmt.Println(p2playMigration)
	fmt.Println("ok")
}