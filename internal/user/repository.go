package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID                 string
	FullName           string
	Phone              *string
	Email              *string
	PasswordHash       string
	Role               string
	MustChangePassword bool
	IsActive           bool
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			full_name,
			phone,
			email,
			password_hash,
			role,
			must_change_password,
			is_active
		FROM users
		WHERE LOWER(email) = LOWER($1)
		`,
		email,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) Create(
	ctx context.Context,
	fullName string,
	email string,
	passwordHash string,
	role string,
	mustChangePassword bool,
) (*User, error) {

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			full_name,
			email,
			password_hash,
			role,
			must_change_password
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			full_name,
			phone,
			email,
			password_hash,
			role,
			must_change_password,
			is_active
		`,
		fullName,
		email,
		passwordHash,
		role,
		mustChangePassword,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) ChangePassword(
	ctx context.Context,
	userID string,
	passwordHash string,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE users
		SET
			password_hash = $1,
			must_change_password = FALSE,
			updated_at = NOW()
		WHERE id = $2
		`,
		passwordHash,
		userID,
	)

	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	userID string,
) (*User, error) {

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			full_name,
			phone,
			email,
			password_hash,
			role,
			must_change_password,
			is_active
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"find user by id: %w",
			err,
		)
	}

	return user, nil
}

func (r *Repository) FindMemberByPhone(
	ctx context.Context,
	phone string,
) (*User, error) {

	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			full_name,
			phone,
			email,
			password_hash,
			role,
			must_change_password,
			is_active
		FROM users
		WHERE phone = $1
		  AND role = 'MEMBER'
		`,
		phone,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.Phone,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.MustChangePassword,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"find member by phone: %w",
			err,
		)
	}

	return user, nil
}
