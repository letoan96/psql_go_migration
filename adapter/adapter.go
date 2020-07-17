package adapter

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/fatih/color"
	_ "github.com/lib/pq"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Connection struct {
	*sql.DB
	*Adapter
}

type Adapter struct {
	Type              string `yaml:"type"`
	Database          string `yaml:"database"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	MaxIdleConnection int    `yaml:"maxIdleConnection"`
	MaxOpenConnection int    `yaml:"maxOpenConnection"`
}

func Initialize(path string, env string) *Adapter {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Can't not read %v  err   #%v ", path, err)
	}

	envConfig := make(map[string]*Adapter)
	err = yaml.Unmarshal(yamlFile, envConfig)

	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	adapter, found := envConfig[env]
	if !found {
		panic(errors.New(fmt.Sprintf(" ========== Can not read configurations of '%s' ᕙ(⇀‸↼‶)ᕗ =========", env)))
	}
	return adapter
}

// Connect to a database with name
func (adapter *Adapter) ConnectToDatabase() *Connection {
	db, err := sql.Open(adapter.Type, fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		adapter.Type,
		adapter.Username,
		adapter.Password,
		adapter.Host,
		adapter.Port,
		adapter.Database))
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(adapter.MaxIdleConnection)
	db.SetMaxOpenConns(adapter.MaxOpenConnection)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	color.Green("Connected to '%s' database at %s:%s\n", adapter.Database, adapter.Host, adapter.Port)
	return &Connection{db, adapter}
}

// Connect to Postgres ONLY ( Then you can create database, run migrations... )
func (adapter *Adapter) ConnectToPostgres() *Connection {
	db, err := sql.Open(adapter.Type, fmt.Sprintf("%s://%s:%s@%s:%s?sslmode=disable",
		adapter.Type,
		adapter.Username,
		adapter.Password,
		adapter.Host,
		adapter.Port))
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(adapter.MaxIdleConnection)
	db.SetMaxOpenConns(adapter.MaxOpenConnection)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println(`Open database connection.`)

	return &Connection{db, adapter}
}

func (c *Connection) Close() {
	if c.DB == nil {
		return
	}

	if err := c.DB.Close(); err != nil {
		color.Red(`Error - Can not close connection`)
	} else {
		color.Yellow(`Database connection closed.`)
	}
}

func (c *Connection) CreatDatabaseIfNotExists() error {
	if c.doesDatabaseExist() == true {
		return errors.New(fmt.Sprintf("Database '%s' already exists.", c.Database))
	} else {
		statement := fmt.Sprintf("CREATE DATABASE %s;", c.Database)
		_, err := c.DB.Exec(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

// The function name has spoken for itself
func (c *Connection) DropDatabase() error {
	if c.doesDatabaseExist() == true {
		statement := fmt.Sprintf("DROP DATABASE %s;", c.Database)
		_, err := c.DB.Exec(statement)
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("Database '%s' does not exists", c.Database))
	}
	return nil
}

func (c *Connection) doesDatabaseExist() bool {
	statement := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", c.Database)
	row := c.DB.QueryRow(statement)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		panic(err)
	}
	return exists
}
