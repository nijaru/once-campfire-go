package database

import (
	"context"
	"database/sql"
	"errors"
)

// PRAGMA data_version is connection-local. Keep one read-only connection and
// prepared statement for the observer, separate from the application's pool.
type versionObserver struct {
	database   *sql.DB
	connection *sql.Conn
	statement  *sql.Stmt
	gate       chan struct{}
}

func openVersionObserver(dsn string) (*versionObserver, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(context.Background())
	if err != nil {
		db.Close()
		return nil, err
	}
	stmt, err := conn.PrepareContext(context.Background(), "PRAGMA data_version")
	if err != nil {
		conn.Close()
		db.Close()
		return nil, err
	}
	return &versionObserver{
		database:   db,
		connection: conn,
		statement:  stmt,
		gate:       make(chan struct{}, 1),
	}, nil
}

func (v *versionObserver) Close() error {
	return errors.Join(v.statement.Close(), v.connection.Close(), v.database.Close())
}

// Observe local and foreign commits without retaining a read transaction between
// calls. Serialize through Scan/Close: overlapping rows can otherwise keep an
// older SQLite snapshot alive even after another connection commits.
func (d *DB) ResponseVersion(ctx context.Context) (uint64, error) {
	v := d.version
	select {
	case v.gate <- struct{}{}:
		defer func() { <-v.gate }()
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	var version uint64
	err := v.statement.QueryRowContext(ctx).Scan(&version)
	return version, err
}
