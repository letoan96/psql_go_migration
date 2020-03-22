package adapter

import (
	"fmt"
	"database/sql"
	"strconv"
	"github.com/fatih/color"
	_ "github.com/lib/pq"

	"gopkg.in/yaml.v2"
    "io/ioutil"
)

type DB struct {
	*sql.DB
	*Adapter
}

type Enviroment struct {
    development Adapter `yaml:"development"`
    staging 	Adapter `yaml:"staging"`
    production  Adapter `yaml:"production"`
}

type Adapter struct {
	Type 		string `yaml:"type"`
	Database    string `yaml:"database"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	MaxIdleConnection int `yaml:"maxIdleConnection"`
	MaxOpenConnection int `yaml:"maxOpenConnection"`
}

func Initialize(path string) {
	yamlFile, err := ioutil.ReadFile(path)
	var en Enviroment
	if err != nil {
        fmt.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        fmt.Fatalf("Unmarshal: %v", err)
    }

    return c
	// maxIdleConnection, err := strconv.Atoi(config["maxIdleConnection"])
	// if err != nil {
	// 	panic(err)
	// }

	// maxOpenConnection, err := strconv.Atoi(config["maxOpenConnection"])
	// if err != nil {
	// 	panic(err)
	// }

	// adapter := Adapter {
	// 	Type: 		 config["type"] ,
	// 	Database:	 config["database"],
	// 	Username: 	 config["username"],
	// 	Password:	 config["password"],
	// 	Host:		 config["host"],
	// 	Port:		 config["port"],
	// 	MaxIdleConnection: maxIdleConnection,
	// 	MaxOpenConnection: maxOpenConnection,
	// }
	// return adapter
}

// type Adapter struct {
// 	Type 		string
// 	Database    string
// 	Username    string
// 	Password    string
// 	Host        string
// 	Port        string
// 	MaxIdleConnection int
// 	MaxOpenConnection int 
// }

// func Initialize(config map[string]string) Adapter {

// 	maxIdleConnection, err := strconv.Atoi(config["maxIdleConnection"])
// 	if err != nil {
// 		panic(err)
// 	}

// 	maxOpenConnection, err := strconv.Atoi(config["maxOpenConnection"])
// 	if err != nil {
// 		panic(err)
// 	}

// 	adapter := Adapter {
// 		Type: 		 config["type"] ,
// 		Database:	 config["database"],
// 		Username: 	 config["username"],
// 		Password:	 config["password"],
// 		Host:		 config["host"],
// 		Port:		 config["port"],
// 		MaxIdleConnection: maxIdleConnection,
// 		MaxOpenConnection: maxOpenConnection,
// 	}
// 	return adapter
// }


// Connect to a database with name
func (adapter *Adapter) ConnectToDatabase() (dbObject *DB, err error) {
	db, err := sql.Open(adapter.Type, fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		adapter.Type,
		adapter.Username,
		adapter.Password,
		adapter.Host,
		adapter.Port,
		adapter.Database))
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(adapter.MaxIdleConnection)
	db.SetMaxOpenConns(adapter.MaxOpenConnection)
	err = db.Ping()
	if err != nil {
		red := color.New(color.FgRed).PrintfFunc()
		red("%s\n", err)
		return nil, err
	}
	color.Green("Connected to '%s' database at %s:%s\n", adapter.Database, adapter.Host, adapter.Port)
	return &DB{db, adapter}, nil
}

// Connect to Postgres ONLY ( Then you can create database, run migrations... )
func (adapter *Adapter) ConnectToPostgres() (dbObject *DB, err error) {
	db, err := sql.Open(adapter.Type, fmt.Sprintf("%s://%s:%s@%s:%s?sslmode=disable",
		adapter.Type,
		adapter.Username,
		adapter.Password,
		adapter.Host,
		adapter.Port))
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(adapter.MaxIdleConnection)
	db.SetMaxOpenConns(adapter.MaxOpenConnection)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println(`Database connection opened.`)
	return &DB{db, adapter}, nil
}

func (c *DB) Close() {
	if c.DB == nil {
		return
	}

	if err := c.DB.Close(); err != nil {
		color.Red(`Error - Can not close connection`)
	} else {
		color.Yellow(`Database connection closed.`)
	}

	return
}