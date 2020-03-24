package adapter
// TODO: Dynamically set enviroment base on any value in database.yml ( currently, We only allow developement, Staging, Production, Test, LocalTest)
import (
	"fmt"
	"database/sql"
	"github.com/fatih/color"
	_ "github.com/lib/pq"

	"gopkg.in/yaml.v2"
    "io/ioutil"
)

type DB struct {
	*sql.DB
	*Adapter
}

type EnviromentConfig struct {
    Development Adapter `yaml:"development"`
    Staging 	Adapter `yaml:"staging"`
    Production  Adapter `yaml:"production"`
    Test 		Adapter `yaml:"test"`
    LocalTest 	Adapter `yaml:"local_test"`
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

func Initialize(path string, env string) Adapter {
	yamlFile, err := ioutil.ReadFile(path)
	envConfig := EnviromentConfig{}
	if err != nil {
        fmt.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, &envConfig)
    if err != nil {
        fmt.Printf("Unmarshal: %v", err)
    }

    switch env {
	    case "development":
	        return envConfig.Development
	    case "staging":
	        return envConfig.Staging
	    case "production":
	    	return envConfig.Production
	    case "test":
	    	return envConfig.Test
	    case "local_test":
	    	return envConfig.LocalTest
	}
	return envConfig.Development
}

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
		return nil, err
	}
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
	return &DB{db, adapter}, nil
}

// The function name has spoken for itself 
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