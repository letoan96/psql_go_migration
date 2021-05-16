package psql_go_migration

import (
	"database/sql"

	"github.com/fatih/color"
	"github.com/letoan96/psql_go_migration/adapter"
	"github.com/letoan96/psql_go_migration/migration"
)

// configFilePath -> path to db config file (database.yaml) file
// migrationDirectoryPath -> folder contains migrate files
// step -> how many migrations will be rolled back?
var red = color.New(color.FgRed).PrintfFunc()
var green = color.New(color.FgGreen).PrintfFunc()
var yellow = color.New(color.FgYellow).PrintfFunc()

func ConnectDB(configFilePath string, environment string) *sql.DB {
	adapter := adapter.Initialize(configFilePath, environment)
	conn := adapter.ConnectToDatabase()
	return conn.DB
}

func CreateDb(configFilePath string, environment string) error {
	adapter := adapter.Initialize(configFilePath, environment)
	conn := adapter.ConnectToPostgres()
	return create(conn)
}

func DropDb(configFilePath string, environment string) error {
	adapter := adapter.Initialize(configFilePath, environment)
	conn := adapter.ConnectToPostgres()
	defer conn.Close()
	return drop(conn)
}

func NewMigration(name string, migrationDirectoryPath string) (string, string) {
	return migration.Generate(name, migrationDirectoryPath)
}

func Migrate(configFilePath string, migrationDirectoryPath string, environment string) {
	green(">> Migrate '%s' database. \n", environment)
	adapter := adapter.Initialize(configFilePath, environment)
	conn := adapter.ConnectToDatabase()
	defer conn.Close()
	migrate(conn.DB, migrationDirectoryPath, adapter.TaskCMD)
}

func Rollback(configFilePath string, migrationDirectoryPath string, environment string, step int) {
	adapter := adapter.Initialize(configFilePath, environment)
	conn := adapter.ConnectToDatabase()
	defer conn.Close()
	rollback(migrationDirectoryPath, conn.DB, step)
}

// Databases --------------------------- For multiple databases -------------------

func ConnectMultipleDB(configFilePath string, environment string, dbName []string) map[string]*sql.DB {
	databases := map[string]*sql.DB{}
	adapterMap := adapter.InitializeMultipleAdapter(configFilePath, environment, dbName)
	for name, adapter := range adapterMap {
		conn := adapter.ConnectToDatabase()
		databases[name] = conn.DB
	}

	return databases
}

func MigrateSingleDB(configFilePath string, migrationDirectoryPath string, environment string, dbName string) {
	green(">> Migrate '%s' database.", environment)
	adapter := adapter.InitializeMultipleAdapter(configFilePath, environment, []string{dbName})[dbName]
	conn := adapter.ConnectToDatabase()
	defer conn.Close()
	migrate(conn.DB, migrationDirectoryPath, adapter.TaskCMD)
}

func CreateSingleDB(configFilePath string, environment string, dbName string) {
	adapter := adapter.InitializeMultipleAdapter(configFilePath, environment, []string{dbName})[dbName]
	conn := adapter.ConnectToPostgres()
	defer conn.Close()
	create(conn)
}

func RollbackSingleDB(configFilePath string, migrationDirectoryPath string, environment string, dbName string, step int) {
	adapter := adapter.InitializeMultipleAdapter(configFilePath, environment, []string{dbName})[dbName]
	conn := adapter.ConnectToDatabase()
	defer conn.Close()
	rollback(migrationDirectoryPath, conn.DB, step)
}

func DropSingleDB(configFilePath string, environment string, dbName string) {
	adapter := adapter.InitializeMultipleAdapter(configFilePath, environment, []string{dbName})[dbName]
	conn := adapter.ConnectToPostgres()
	defer conn.Close()

	drop(conn)
}

//-----------------------------------------
func create(connection *adapter.Connection) error {
	defer connection.Close()
	return connection.CreateDatabaseIfNotExists()

}

func migrate(db *sql.DB, migrationPath string, taskCMD string) {
	migration := migration.Initialize(db, migrationPath, taskCMD)
	migration.Migrate()
}

func rollback(migrationDirectoryPath string, db *sql.DB, step int) {
	migration := migration.Initialize(db, migrationDirectoryPath)
	migration.RollBack(step)
}

func drop(conn *adapter.Connection) error {
	return conn.DropDatabase()
}
