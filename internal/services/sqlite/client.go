package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Client wraps a SQLite database connection
type Client struct {
	db *sql.DB
}

// Result mimics sql.Result
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// NewClient creates a new SQLite client from a path
func NewClient(path string) (*Client, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable useful PRAGMAs
	_, _ = db.Exec("PRAGMA foreign_keys = ON")
	_, _ = db.Exec("PRAGMA journal_mode = WAL")

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// Query executes a query and returns map results
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMaps(rows)
}

// Exec executes a statement
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}

	lastID, _ := res.LastInsertId()
	rows, _ := res.RowsAffected()
	return Result{LastInsertID: lastID, RowsAffected: rows}, nil
}

// Begin starts a transaction
func (c *Client) Begin(ctx context.Context) (*Tx, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Tx wraps a SQL transaction
type Tx struct {
	tx *sql.Tx
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMaps(rows)
}

func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return Result{}, err
	}

	lastID, _ := res.LastInsertId()
	rows, _ := res.RowsAffected()
	return Result{LastInsertID: lastID, RowsAffected: rows}, nil
}

// Helper to convert rows to map slice
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]
			if b, ok := val.([]byte); ok {
				row[colName] = string(b)
			} else {
				row[colName] = val
			}
		}
		results = append(results, row)
	}

	return results, rows.Err()
}
