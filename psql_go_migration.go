package psql_go_migration

import (
	"database/sql"
	"fmt"

	"github.com/fatih/color"
	"github.com/letoan96/psql_go_migration/adapter"
	"github.com/letoan96/psql_go_migration/migration"
)

// configFilePath -> path to db config file (database.yaml) file
// enviroment -> development or test or staging ?
// migrationDirectoryPath -> folder contains migration files
// step -> how many migrations will be rolled back?
//

func ConnectDB(configFilePath string, enviroment string) *sql.DB {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	color.Red("========================   Try `lottery_tools -db migrate` when encounter weird errors ======================")
	connection := connect(adapterInstance)
	return connection.DB
}

func CreateDb(configFilePath string, enviroment string) {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	psql := adapterInstance.ConnectToPostgres()
	create(psql)
}

func DropDb(configFilePath string, enviroment string) {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	psql := adapterInstance.ConnectToPostgres()
	defer psql.Close()

	err := psql.DropDatabase()
	if err != nil {
		red := color.New(color.FgRed).PrintfFunc()
		red("%s\n", err)
		return
	}

	color.Yellow(`Database '%s' droped.`, psql.Database)
}

func NewMigration(name string, migrationDirectoryPath string) {
	migration.Generate(name, migrationDirectoryPath)
}

func MigrateDb(configFilePath string, migrationDirectoryPath string, enviroment string) {
	fmt.Println(">> Migrate ", enviroment, "database")
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	connection := adapterInstance.ConnectToDatabase()

	migrate(connection.DB, migrationDirectoryPath)
}

func Rollback(configFilePath string, migrationDirectoryPath string, enviroment string, step int) {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	connection := adapterInstance.ConnectToDatabase()
	rollback(migrationDirectoryPath, connection.DB, step)
	//migrationInstance := migration.Initialize(connection.DB, migrationDirectoryPath)
	//migrationInstance.RollBack(step)

}

// ------------------------------------------------------------------- For mutiple databases -------------------
// These functions for mutiple database
type Databases map[string]*sql.DB

func ConnectMutipleDB(configFilePath string, enviroment string, dbName []string) (databases Databases) {
	adapters := adapter.InitializeMutipleAdapter(configFilePath, enviroment, dbName)

	for i, a := range adapters {
		connection := connect(a)
		name := dbName[i]
		databases[name] = connection.DB
	}

	return databases
}

func MigrateSingleDB(configFilePath string, migrationDirectoryPath string, enviroment string, dbName string) {
	fmt.Println(fmt.Sprintf(`>> Migrate '%s' '%s' database`, enviroment, dbName))
	adapterInstance := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := adapterInstance.ConnectToDatabase()

	migrate(connection.DB, migrationDirectoryPath)
}

func CreateSingleDB(configFilePath string, enviroment string, dbName string) {
	adapter := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := connect(adapter)
	create(connection)
}

func RollbackSingleDB(configFilePath string, migrationDirectoryPath string, enviroment string, dbName string, step int) {
	adapterInstance := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := adapterInstance.ConnectToDatabase()

	rollback(migrationDirectoryPath, connection.DB, step)
}

func DropSingleDB(configFilePath string, enviroment string, dbName string) {
	adapterInstance := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := adapterInstance.ConnectToPostgres()
	drop(connection)

}

//-----------------------------------------
func connect(a *adapter.Adapter) *adapter.Connection {
	connection := a.ConnectToDatabase()
	connection.DB.Exec(`SET TIMEZONE='Asia/Bangkok';`)
	return connection
}

func create(connection *adapter.Connection) {
	defer connection.Close()
	err := connection.CreatDatabaseIfNotExists()
	if err != nil {
		red := color.New(color.FgRed).PrintfFunc()
		red("%s\n", err)
		return
	}

	color.Green(`Created '%s' database.`, connection.Database)
}

func migrate(db *sql.DB, migrationPath string) {
	migrationInstance := migration.Initialize(db, migrationPath)
	migrationInstance.Migrate()
}

func rollback(migrationDirectoryPath string, db *sql.DB, step int) {
	migrationInstance := migration.Initialize(db, migrationDirectoryPath)
	migrationInstance.RollBack(step)
}

func drop(connection *adapter.Connection) {
	defer connection.Close()

	err := connection.DropDatabase()
	if err != nil {
		red := color.New(color.FgRed).PrintfFunc()
		red("%s\n", err)
		return
	}

	color.Yellow(`Database '%s' droped.`, connection.Database)
}
