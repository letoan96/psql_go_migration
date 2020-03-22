package adapter

import (
	"fmt"
	"database/sql"
	"strconv"
	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func Test() {
	fmt.Println("package adapter")
}

type Adapter struct {
	Type 		string
	Database    string
	Username    string
	Password    string
	Host        string
	Port        string
	MaxIdleConnection int
	MaxOpenConnection int 
}
// || "postgresql"
func Initialize(config map[string]string) Adapter {
	maxIdleConnection, err := strconv.Atoi(config["maxIdleConnection"])
	if err != nil {
		panic(err)
	}

	maxOpenConnection, err := strconv.Atoi(config["maxOpenConnection"])
	if err != nil {
		panic(err)
	}

	adapter := Adapter {
		Type: 		 config["type"] ,
		Database:	 config["database"],
		Username: 	 config["username"],
		Password:	 config["password"],
		Host:		 config["host"],
		Port:		 config["port"],
		MaxIdleConnection: maxIdleConnection,
		MaxOpenConnection: maxOpenConnection,
	}
	return adapter
}

func (adapter *Adapter) Connect() (dbObject *DB, err error) {
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
	return &DB{db}, nil
}