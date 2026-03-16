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
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(dsn string) error {
	if dsn == "" {
		dsn = "postgres://user:password@localhost/code_together?sslmode=disable"
	}

	dbsql, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer dbsql.Close()

	driver, err := pgx.WithInstance(dbsql, &pgx.Config{})
	if err != nil {
		return err
	}

	src := migrationsPath()
	m, err := migrate.NewWithDatabaseInstance(src, "pgx5", driver)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}
	return nil
}

func migrationsPath() string {
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "server", "migrations"),
		filepath.Join(cwd, "migrations"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return fmt.Sprintf("file://%s", p)
		}
	}
	return "file://migrations"
}
