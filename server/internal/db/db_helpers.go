// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB struct wraps the Queries struct with a connection pool
// This provides the database connection functionality used by repositories
type DB struct {
	*Queries
	connPool *pgxpool.Pool
}

// ConnectDB creates a new database connection pool
func ConnectDB(dsn string) (*DB, error) {
	if dsn == "" {
		dsn = "postgres://user:password@localhost/code_together?sslmode=disable"
	}

	connPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	queries := &Queries{db: connPool}

	return &DB{
		Queries:  queries,
		connPool: connPool,
	}, nil
}

// Close closes the database connection pool
func (d *DB) Close() {
	d.connPool.Close()
}

// Pool returns the underlying pgxpool.Pool
func (d *DB) Pool() *pgxpool.Pool {
	return d.connPool
}

// Conn returns the underlying pgxpool.Pool for raw queries
func (d *DB) Conn() *pgxpool.Pool {
	return d.connPool
}

// DB returns the underlying database connection (for use in legacy code that needs sql.DB)
func (d *DB) DB() *pgxpool.Pool {
	return d.connPool
}
