package adapter

type Connectable interface {
	ConnectToDatabase() *Connection
	ConnectToPostgres() *Connection
}

func (conn *Connection) Close()                           { conn.close() }
func (conn *Connection) CreateDatabaseIfNotExists() error { return conn.createDatabaseIfNotExists() }
func (conn *Connection) DropDatabase() error              { return conn.dropDatabase() }
