package psql_go_migration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/letoan96/psql_go_migration/adapter"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	c := NewConfig("development", "./db/migrate.yaml", "./db/database.yaml")
	c.DropDatabase()
	code := m.Run()
	os.Exit(code)
}

func TestCreateDropDatabase(t *testing.T) {
	c := NewConfig("development", "./db/migrate", "./db/database.yaml")

	err := c.CreateDatabase()
	assert.Equal(t, nil, err, "")

	err = c.CreateDatabase()
	assert.Equal(t, errors.New(adapter.ErrDatabaseAlreadyExists), err, "")

	err = c.DropDatabase()
	assert.Equal(t, nil, err, "")

	err = c.DropDatabase()
	assert.Equal(t, errors.New(adapter.ErrDatabaseDoesNotExist), err, "")
}

func TestConnectDatabase(t *testing.T) {
	c := NewConfig("development", "./db/migrate", "./db/database.yaml")
	defer c.DropDatabase()

	err := c.CreateDatabase()
	assert.Equal(t, nil, err, "")

	db := c.ConnectDatabase()
	assert.NotEqual(t, nil, db, "")

	assert.Panics(t, func() { c.DropDatabase() })

	db.Close()
}

func TestCreateMigration(t *testing.T) {
	c := NewConfig("development", "./db/migrate", "./db/database.yaml")
	assert.NotPanics(t, func() { c.NewMigration("create_user") })

	dir, _ := ioutil.ReadDir("./db/migrate")
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{"migrate", d.Name()}...))
	}
}

func TestMigrate(t *testing.T) {
	defer cleanMigrateFolder()

	c := NewConfig("development", "./db/migrate", "./db/database.yaml")
	c.DropDatabase()
	c.CreateDatabase()

	up, down := c.NewMigration("create_user")

	qUp := fmt.Sprintf(`CREATE TABLE users (username text);`)
	fUp := fmt.Sprintf("./db/migrate/%s", up)
	if err := ioutil.WriteFile(fUp, []byte(qUp), 0644); err != nil {
		panic(err)
	}

	qDown := fmt.Sprintf(`DROP TABLE users;`)
	fDown := fmt.Sprintf("./db/migrate/%s", down)
	if err := ioutil.WriteFile(fDown, []byte(qDown), 0644); err != nil {
		panic(err)
	}

	c.MigrateDatabase()
	c.Rollback(1)
}

func TestMultipleDatabase(t *testing.T) {
	defer cleanMigrateFolder()
	databases := []string{"primary", "replica", "replica2"}
	c := NewConfigMultipleDatabase("development", "./db/migrate", "./db/multiple_database.yaml", databases)
	c.CreateDatabase("primary")
	c.CreateDatabase("replica")
	c.CreateDatabase("replica2")

	db := c.ConnectDatabase()
	assert.Panics(t, func() { c.DropDatabase("primary") })
	assert.Panics(t, func() { c.DropDatabase("replica") })
	assert.Panics(t, func() { c.DropDatabase("replica2") })

	up, down := c.NewMigration("create_user")
	qUp := fmt.Sprintf(`CREATE TABLE users (username text);`)
	fUp := fmt.Sprintf("./db/migrate/%s", up)
	if err := ioutil.WriteFile(fUp, []byte(qUp), 0644); err != nil {
		panic(err)
	}

	qDown := fmt.Sprintf(`DROP TABLE users;`)
	fDown := fmt.Sprintf("./db/migrate/%s", down)
	if err := ioutil.WriteFile(fDown, []byte(qDown), 0644); err != nil {
		panic(err)
	}

	c.MigrateDatabase("primary")
	c.Rollback("primary", 1)

	db["primary"].Close()
	db["replica"].Close()
	db["replica2"].Close()

	c.DropDatabase("primary")
	c.DropDatabase("replica")
	c.DropDatabase("replica2")
}

func cleanMigrateFolder() {
	dir, _ := ioutil.ReadDir("./db/migrate")
	for _, d := range dir {
		os.RemoveAll(fmt.Sprintf(`%s/%s`, "./db/migrate", d.Name()))
	}
}
