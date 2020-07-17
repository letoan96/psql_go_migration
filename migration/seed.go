package migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
)

func Seed(db *sql.DB, path string) error {
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
