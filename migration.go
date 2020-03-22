package migration

import (
	"fmt"
	"github.com/letoan96/psql_go_migration/adapter"
)

adapter.Test()

func (db *adapter.DB) migrate() {
	fmt.Println(`It works`)
}