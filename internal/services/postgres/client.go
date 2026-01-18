package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Client wraps a PostgreSQL database connection
type Client struct {
	db *sql.DB
}

// ConnectOptions holds the connection configuration
type ConnectOptions struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// Result mimics sql.Result
type Result struct {
	RowsAffected int64
}

// NewClient creates a new Postgres client from a connection string
func NewClient(connStr string) (*Client, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db: db}, nil
}

// NewClientFromOptions creates a new Postgres client from options
func NewClientFromOptions(opts ConnectOptions) (*Client, error) {
	// Defaults
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 5432
	}
	if opts.User == "" {
		opts.User = "postgres"
	}
	if opts.Database == "" {
		opts.Database = "postgres"
	}
	if opts.SSLMode == "" {
		opts.SSLMode = "disable"
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.User, opts.Password, opts.Database, opts.SSLMode)

	return NewClient(connStr)
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

	rows, _ := res.RowsAffected()
	return Result{RowsAffected: rows}, nil
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

	rows, _ := res.RowsAffected()
	return Result{RowsAffected: rows}, nil
}

// Helper to convert rows to map slice
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Prepare a slice of interface{} to scan into
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create map for this row
		row := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]

			// Convert []byte to string for easier Lua handling
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
