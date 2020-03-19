package psql_go_migration

import (
	"fmt"
	"github.com/letoan96/psql_go_migration/adapter"
)

func main() {
	fmt.Println(`hello`)
	adapter.Test("hello from the other side")
}