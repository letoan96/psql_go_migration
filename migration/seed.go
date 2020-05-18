package migration

import (
	"fmt"
	"io/ioutil"
	"github.com/letoan96/psql_go_migration/adapter"
)

func Seed(db *adapter.DB, path string) error {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	statement := string(dat)
	fmt.Println("========== Seeding ==================")
	_, err = db.Exec(statement)
	if err != nil {
		return err
	}
	fmt.Println("========== Done ====================")
	
	return nil
}