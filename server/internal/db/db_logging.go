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
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// loggingDB wraps a DBTX to log SQL queries when debug mode is enabled
type loggingDB struct {
	db    DBTX
	debug bool
}

// newLoggingDB creates a new logging wrapper around a DBTX
func newLoggingDB(db DBTX, debug bool) *loggingDB {
	return &loggingDB{
		db:    db,
		debug: debug,
	}
}

func (l *loggingDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()
	if l.debug {
		log.Printf("[SQL] Exec: %s Args: %v", sql, args)
	}
	result, err := l.db.Exec(ctx, sql, args...)
	if l.debug {
		duration := time.Since(start)
		if err != nil {
			log.Printf("[SQL] Exec error (%v): %s", duration, err)
		} else {
			log.Printf("[SQL] Exec success (%v): %s", duration, result)
		}
	}
	return result, err
}

func (l *loggingDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()
	if l.debug {
		log.Printf("[SQL] Query: %s Args: %v", sql, args)
	}
	result, err := l.db.Query(ctx, sql, args...)
	if l.debug {
		duration := time.Since(start)
		if err != nil {
			log.Printf("[SQL] Query error (%v): %s", duration, err)
		} else {
			log.Printf("[SQL] Query success (%v)", duration)
		}
	}
	return result, err
}

func (l *loggingDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()
	if l.debug {
		log.Printf("[SQL] QueryRow: %s Args: %v", sql, args)
	}
	result := l.db.QueryRow(ctx, sql, args...)
	if l.debug {
		duration := time.Since(start)
		log.Printf("[SQL] QueryRow sent (%v)", duration)
	}
	return result
}

// ConnectDBWithDebug creates a new database connection pool with optional SQL logging
func ConnectDBWithDebug(dsn string, debug bool) (*DB, error) {
	if dsn == "" {
		dsn = "postgres://user:password@localhost/code_together?sslmode=disable"
	}

	connPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	var dbtx DBTX = connPool
	if debug {
		dbtx = newLoggingDB(connPool, debug)
		log.Println("[SQL Debug mode enabled]")
	}

	queries := &Queries{db: dbtx}

	return &DB{
		Queries:  queries,
		connPool: connPool,
	}, nil
}
