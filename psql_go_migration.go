package psql_go_migration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/letoan96/psql_go_migration/adapter"
	"github.com/letoan96/psql_go_migration/migration"
)

// configFilePath -> path to db config file (database.yaml) file
func ConnectDB(configFilePath string, enviroment string) *sql.DB {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	// connection.DB.Exec(`SET TIMEZONE='Asia/Bangkok';`)
	color.Red("========================   Try `lottery_tools -db migrate` when encounter weird errors ======================")
	connection := connect(adapterInstance)
	return connection.DB
}

func connect(a *adapter.Adapter) *adapter.Connection {
	connection := a.ConnectToDatabase()
	connection.DB.Exec(`SET TIMEZONE='Asia/Bangkok';`)
	return connection
}

func CreateDb(configFilePath string, enviroment string) {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	psql := adapterInstance.ConnectToPostgres()
	create(psql)
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

func NewMigration(name *string, migrationDirectoryPath string) {
	migration.Generate(*name, migrationDirectoryPath)
}

func MigrateDb(configFilePath string, migrationDirectoryPath string, enviroment string) {
	fmt.Println(">> Migrate ", enviroment, "database")
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	connection := adapterInstance.ConnectToDatabase()

	migrate(connection.DB, migrationDirectoryPath)
}

func migrate(db *sql.DB, migrationPath string) {
	migrationInstance := migration.Initialize(db, migrationPath)
	migrationInstance.Migrate()
}

func Rollback(configFilePath string, migrationDirectoryPath string, enviroment string, step int) {
	adapterInstance := adapter.Initialize(configFilePath, enviroment)
	connection := adapterInstance.ConnectToDatabase()
	migrationInstance := migration.Initialize(connection.DB, migrationDirectoryPath)
	migrationInstance.RollBack(step)
}

// ------------------------------------------------------------------- For mutiple databases -------------------

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
	panic(errors.New("asdasdas"))
	adapterInstance := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := adapterInstance.ConnectToDatabase()

	migrate(connection.DB, migrationDirectoryPath)
}

func CreateSingleDB(configFilePath string, enviroment string, dbName string) {
	adapter := adapter.InitializeAdapter(configFilePath, enviroment, dbName)
	connection := connect(adapter)
	create(connection)
}

// var seedName = flag.String("seedname", "", "Data seed name")

// func Seed() {
// 	var env *ENV
// 	renv.ParseCmd(&env)

// 	adapterInstance := adapter.Initialize(env.DatabaseConfigFilePath, env.Enviroment)
// 	connection := adapterInstance.ConnectToDatabase()

// 	var seedFile string
// 	if *seedName == "" {
// 		seedFile = fmt.Sprintf("%s/seed.sql", env.DatabaseSeedFilePath)
// 	} else {
// 		seedFile = fmt.Sprintf("%s/%s.sql", env.DatabaseSeedFilePath, *seedName)
// 	}

// 	fmt.Printf("Running seed file: %s\n", seedFile)

// 	err := migration.Seed(connection.DB, seedFile)
// 	if err != nil {
// 		red := color.New(color.FgRed).PrintfFunc()
// 		red("%s\n", err)
// 		panic(err)
// 	}
// }
