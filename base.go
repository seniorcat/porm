package porm

import "database/sql"

// Porm general mod interface
type Porm struct {
	db IDb
	q  queryStruct
}

// IDb interface for get a new instanse of database
type IDb interface {
	New() DS
}

// DS database interface for run some query
type DS interface {
	Query(string) []map[string]string
	QueryRow(string) map[string]string
	Count(string) int64
	Begin() (*sql.Tx, error)
	PrepareInsert(tx *sql.Tx, q string) (stmt *sql.Stmt)
	Exec(stmt *sql.Stmt, f ...interface{}) (count int64)
}

// GetDb method for get db instanse
func (p *Porm) GetDb() DS {
	return p.db.New()
}
