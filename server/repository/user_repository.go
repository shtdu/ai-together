// Copyright (c) 2025 AI Together
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

package repository

import (
	"context"
	"fmt"
	"time"

	"switch-server/internal/db"
	"switch-server/models"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(database *db.DB) *UserRepository {
	return &UserRepository{
		db: database,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, password, name, role string, tenantID int64) (*models.User, error) {
	// Insert the user into the database (password is already hashed by the service)
	err := r.db.CreateUser(ctx, db.CreateUserParams{
		Email:    email,
		Name:     name,
		Password: password,
		Role:     role,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get the user by email to return the full user object
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	createdAt := time.Time{}
	if user.CreatedAt.Valid {
		createdAt = user.CreatedAt.Time
	}

	updatedAt := time.Time{}
	if user.UpdatedAt.Valid {
		updatedAt = user.UpdatedAt.Time
	}

	model := &models.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.Password,
		Role:      user.Role,
		TenantID:  user.TenantID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return model, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	user, err := r.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	model := &models.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.Password,
		Role:      user.Role,
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}

	return model, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, userID int64, name, role string) error {
	err := r.db.UpdateUser(ctx, db.UpdateUserParams{
		ID:   userID,
		Name: name,
		Role: role,
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateUserWithPassword(ctx context.Context, userID int64, name, role, password string) error {
	query := `UPDATE users SET name = $2, role = $3, password = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Pool().Exec(ctx, query, userID, name, role, password)
	if err != nil {
		return fmt.Errorf("failed to update user with password: %w", err)
	}
	return nil
}

func (r *UserRepository) ListUsersByTenant(ctx context.Context, tenantID int64) ([]models.User, error) {
	query := `SELECT id, email, name, password, role, tenant_id, created_at, updated_at FROM users WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Pool().Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by tenant: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var createdAt, updatedAt time.Time
		err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.Role, &u.TenantID, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) UpdateUserName(ctx context.Context, userID int64, name string) error {
	query := `UPDATE users SET name = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Pool().Exec(ctx, query, userID, name)
	if err != nil {
		return fmt.Errorf("failed to update user name: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateUserPassword(ctx context.Context, userID int64, hashedPassword string) error {
	query := `UPDATE users SET password = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Pool().Exec(ctx, query, userID, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Pool().Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
