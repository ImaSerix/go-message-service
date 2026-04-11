package query

import "database/sql"

func NewQueryService(db *sql.DB, limit int) *QueryService {
	return &QueryService{db: db, limit: limit}
}
