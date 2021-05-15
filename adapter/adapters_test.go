package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeMultipleAdapter(t *testing.T) {
	dbConfigPath := "./test/multiple_database.yaml"

	db := InitializeMultipleAdapter(dbConfigPath, "development", nil)
	assert.Equal(t, len(db), 3, "")

	db = InitializeMultipleAdapter(dbConfigPath, "development", []string{"primary", "replica"})
	assert.Equal(t, len(db), 2, "")
	_, found := db["primary"]
	assert.Equal(t, found, true, "")

	_, found = db["replica"]
	assert.Equal(t, found, true, "")

	_, found = db["replica2"]
	assert.Equal(t, found, false, "")
}

func TestInitialize(t *testing.T) {
	dbConfigPath := "./test/database.yaml"

	db := Initialize(dbConfigPath, "development")
	assert.Equal(t, db.Type, "postgres", "")
	assert.Equal(t, db.Database, "development_database_test", "")
	assert.Equal(t, db.Host, "127.0.0.1", "")
	assert.Equal(t, db.Port, "5433", "")
}
