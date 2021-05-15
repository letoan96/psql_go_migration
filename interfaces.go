package psql_go_migration

import "database/sql"

type Config struct {
	Environment              string // environment -> development or db or staging ?
	PathToMigrationDirectory string // migrationDirectoryPath -> the folder that contains migrate files
	PathToDatabaseConfigFile string //path to db config file (database.yaml) file
}

func NewConfig(env, pathToMigration, pathToConfigFile string) *Config {
	return &Config{Environment: env, PathToMigrationDirectory: pathToMigration, PathToDatabaseConfigFile: pathToConfigFile}
}

func (c *Config) ConnectDatabase() *sql.DB {
	return ConnectDB(c.PathToDatabaseConfigFile, c.Environment)
}

func (c *Config) CreateDatabase() error { return CreateDb(c.PathToDatabaseConfigFile, c.Environment) }

func (c *Config) DropDatabase() error { return DropDb(c.PathToDatabaseConfigFile, c.Environment) }

func (c *Config) NewMigration(name string) (string, string) {
	return NewMigration(name, c.PathToMigrationDirectory)
}

func (c *Config) MigrateDatabase() {
	Migrate(c.PathToDatabaseConfigFile, c.PathToMigrationDirectory, c.Environment)
}

func (c *Config) RollBack(step int) {
	Rollback(c.PathToDatabaseConfigFile, c.PathToMigrationDirectory, c.Environment, step)
}

// Databases --------------------------- For multiple databases -------------------

type ConfigMultipleDatabase struct {
	ListDatabase []string
	*Config
}

func NewConfigMultipleDatabase(env, pathToMigration, pathToConfigFile string, databases []string) *ConfigMultipleDatabase {
	return &ConfigMultipleDatabase{
		databases,
		&Config{Environment: env, PathToMigrationDirectory: pathToMigration, PathToDatabaseConfigFile: pathToConfigFile},
	}
}

func (c *ConfigMultipleDatabase) ConnectDatabase() map[string]*sql.DB {
	return ConnectMultipleDB(c.PathToDatabaseConfigFile, c.Environment, c.ListDatabase)
}

func (c *ConfigMultipleDatabase) CreateDatabase(name string) {
	CreateSingleDB(c.PathToDatabaseConfigFile, c.Environment, name)
}

func (c *ConfigMultipleDatabase) DropDatabase(name string) {
	DropSingleDB(c.PathToDatabaseConfigFile, c.Environment, name)
}

func (c *ConfigMultipleDatabase) MigrateDatabase(name string) {
	MigrateSingleDB(c.PathToDatabaseConfigFile, c.PathToMigrationDirectory, c.Environment, name)
}

func (c *ConfigMultipleDatabase) Rollback(name string, step int) {
	RollbackSingleDB(c.PathToDatabaseConfigFile, c.PathToMigrationDirectory, c.Environment, name, step)
}
