package member

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nati-darse/edirlink-backend/internal/ethiopiancalendar"
)

type Member struct {
	ID                 string  `json:"id"`
	FullName           string  `json:"full_name"`
	Phone              string  `json:"phone"`
	Email              *string `json:"email"`
	Gender             *string `json:"gender"`
	Address            *string `json:"address"`
	MustChangePassword bool    `json:"must_change_password"`
	IsActive           bool    `json:"is_active"`
}

type Repository struct {
	db       *pgxpool.Pool
	calendar *ethiopiancalendar.Service
}

func NewRepository(
	db *pgxpool.Pool,
	calendar *ethiopiancalendar.Service,
) *Repository {
	return &Repository{
		db:       db,
		calendar: calendar,
	}
}

// CheckPhoneInEdir
func (r *Repository) CheckPhoneInEdir(
	ctx context.Context,
	edirID string,
	phone string,
) (bool, error) {

	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM memberships m
			JOIN users u ON u.id = m.member_id
			WHERE m.edir_id = $1
			  AND u.phone = $2
			  AND m.is_active = TRUE
		)
		`,
		edirID,
		phone,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf(
			"check member phone: %w",
			err,
		)
	}

	return exists, nil
}

// CreateUser creates the actual MEMBER user account.
func (r *Repository) CreateUser(
	ctx context.Context,
	fullName string,
	phone string,
	email *string,
	passwordHash string,
) (*Member, error) {

	member := &Member{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			full_name,
			phone,
			email,
			password_hash,
			role,
			must_change_password
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'MEMBER',
			TRUE
		)
		RETURNING
			id,
			full_name,
			phone,
			email,
			must_change_password,
			is_active
		`,
		fullName,
		phone,
		email,
		passwordHash,
	).Scan(
		&member.ID,
		&member.FullName,
		&member.Phone,
		&member.Email,
		&member.MustChangePassword,
		&member.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create member user: %w",
			err,
		)
	}

	return member, nil
}

func (r *Repository) CreateMembership(
	ctx context.Context,
	edirID string,
	memberID string,
	gender *string,
	address *string,
) error {
	today := r.calendar.Today()

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO memberships (
			edir_id,
			member_id,
			gender,
			address,
			joined_ethiopian_year,
			joined_ethiopian_month,
			joined_ethiopian_day
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		edirID,
		memberID,
		gender,
		address,
		today.Year,
		today.Month,
		today.Day,
	)

	if err != nil {
		return fmt.Errorf("create membership: %w", err)
	}

	return nil
}

func (r *Repository) ListMembers(
	ctx context.Context,
	edirID string,
) ([]Member, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			u.id,
			u.full_name,
			u.phone,
			u.email,
			m.gender,
			m.address,
			u.must_change_password,
			u.is_active
		FROM memberships m
		JOIN users u
			ON u.id = m.member_id
		WHERE m.edir_id = $1
		  AND m.is_active = TRUE
		  AND u.role = 'MEMBER'
		ORDER BY u.full_name ASC
		`,
		edirID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list members: %w",
			err,
		)
	}

	defer rows.Close()

	var members []Member

	for rows.Next() {

		var member Member

		err := rows.Scan(
			&member.ID,
			&member.FullName,
			&member.Phone,
			&member.Email,
			&member.Gender,
			&member.Address,
			&member.MustChangePassword,
			&member.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scan member: %w",
				err,
			)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate members: %w",
			err,
		)
	}

	return members, nil
}

func (r *Repository) FindByIDInEdir(
	ctx context.Context,
	edirID string,
	memberID string,
) (*Member, error) {

	member := &Member{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			u.id,
			u.full_name,
			u.phone,
			u.email,
			m.gender,
			m.address,
			u.must_change_password,
			u.is_active
		FROM memberships m
		JOIN users u ON u.id = m.member_id
		WHERE m.edir_id = $1
		  AND m.member_id = $2
		  AND m.is_active = TRUE
		`,
		edirID,
		memberID,
	).Scan(
		&member.ID,
		&member.FullName,
		&member.Phone,
		&member.Email,
		&member.Gender,
		&member.Address,
		&member.MustChangePassword,
		&member.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"find member in edir: %w",
			err,
		)
	}

	return member, nil
}

func (r *Repository) AddExistingMember(
	ctx context.Context,
	edirID string,
	memberID string,
	gender *string,
	address *string,
) error {
	var role string
	var isActive bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT role, is_active
		FROM users
		WHERE id = $1
		`,
		memberID,
	).Scan(&role, &isActive)

	if err != nil {
		return fmt.Errorf("find member: %w", err)
	}

	if role != "MEMBER" {
		return fmt.Errorf("user is not a member")
	}

	if !isActive {
		return fmt.Errorf("member account is inactive")
	}

	today := r.calendar.Today()

	// A new membership, or the reactivation of an inactive one, starts on
	// today's Ethiopian date. An already-active membership keeps its date.
	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO memberships (
			edir_id,
			member_id,
			gender,
			address,
			is_active,
			joined_at,
			joined_ethiopian_year,
			joined_ethiopian_month,
			joined_ethiopian_day
		)
		VALUES ($1, $2, $3, $4, TRUE, NOW(), $5, $6, $7)
		ON CONFLICT (edir_id, member_id)
		DO UPDATE SET
			is_active = TRUE,
			gender = EXCLUDED.gender,
			address = EXCLUDED.address,
			joined_at = CASE
				WHEN memberships.is_active = FALSE
				THEN NOW()
				ELSE memberships.joined_at
			END,
			joined_ethiopian_year = CASE
				WHEN memberships.is_active = FALSE
				THEN EXCLUDED.joined_ethiopian_year
				ELSE memberships.joined_ethiopian_year
			END,
			joined_ethiopian_month = CASE
				WHEN memberships.is_active = FALSE
				THEN EXCLUDED.joined_ethiopian_month
				ELSE memberships.joined_ethiopian_month
			END,
			joined_ethiopian_day = CASE
				WHEN memberships.is_active = FALSE
				THEN EXCLUDED.joined_ethiopian_day
				ELSE memberships.joined_ethiopian_day
			END
		`,
		edirID,
		memberID,
		gender,
		address,
		today.Year,
		today.Month,
		today.Day,
	)

	if err != nil {
		return fmt.Errorf("add existing member: %w", err)
	}

	return nil
}

func (r *Repository) FindByPhone(
	ctx context.Context,
	phone string,
) (*Member, error) {

	member := &Member{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			full_name,
			phone,
			email,
			must_change_password,
			is_active
		FROM users
		WHERE phone = $1
		  AND role = 'MEMBER'
		`,
		phone,
	).Scan(
		&member.ID,
		&member.FullName,
		&member.Phone,
		&member.Email,
		&member.MustChangePassword,
		&member.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"find member by phone: %w",
			err,
		)
	}

	return member, nil
}

func (r *Repository) AddToEdir(
	ctx context.Context,
	edirID string,
	memberID string,
	gender *string,
	address *string,
) error {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM memberships
			WHERE edir_id = $1
			  AND member_id = $2
		)
		`,
		edirID,
		memberID,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}

	today := r.calendar.Today()

	if exists {
		// Reactivation starts fresh: the joining date is overwritten with
		// today's Ethiopian date. An already-active membership keeps its date.
		_, err = r.db.Exec(
			ctx,
			`
			UPDATE memberships
			SET
				is_active = TRUE,
				gender = $3,
				address = $4,
				joined_at = CASE
					WHEN is_active = FALSE THEN NOW()
					ELSE joined_at
				END,
				joined_ethiopian_year = CASE
					WHEN is_active = FALSE THEN $5::INTEGER
					ELSE joined_ethiopian_year
				END,
				joined_ethiopian_month = CASE
					WHEN is_active = FALSE THEN $6::SMALLINT
					ELSE joined_ethiopian_month
				END,
				joined_ethiopian_day = CASE
					WHEN is_active = FALSE THEN $7::SMALLINT
					ELSE joined_ethiopian_day
				END
			WHERE edir_id = $1
			  AND member_id = $2
			`,
			edirID,
			memberID,
			gender,
			address,
			today.Year,
			today.Month,
			today.Day,
		)

		if err != nil {
			return fmt.Errorf("reactivate membership: %w", err)
		}

		return nil
	}

	_, err = r.db.Exec(
		ctx,
		`
		INSERT INTO memberships (
			edir_id,
			member_id,
			gender,
			address,
			joined_ethiopian_year,
			joined_ethiopian_month,
			joined_ethiopian_day
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		edirID,
		memberID,
		gender,
		address,
		today.Year,
		today.Month,
		today.Day,
	)

	if err != nil {
		return fmt.Errorf("add member to edir: %w", err)
	}

	return nil
}

func (r *Repository) RemoveFromEdir(
	ctx context.Context,
	edirID string,
	memberID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`
		UPDATE memberships
		SET is_active = FALSE
		WHERE edir_id = $1
		  AND member_id = $2
		`,
		edirID,
		memberID,
	)

	if err != nil {
		return fmt.Errorf("remove member from edir: %w", err)
	}

	return nil
}

func (r *Repository) ListMyEdirs(
	ctx context.Context,
	memberID string,
) ([]map[string]interface{}, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			e.id,
			e.name,
			e.monthly_contribution_amount,
			e.due_ethiopian_day,
			e.support_amount,
			e.penalty_percentage,
			e.status
		FROM memberships m
		JOIN edirs e ON e.id = m.edir_id
		WHERE m.member_id = $1
		  AND m.is_active = TRUE
		ORDER BY e.name
		`,
		memberID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list member edirs: %w",
			err,
		)
	}

	defer rows.Close()

	edirs := make([]map[string]interface{}, 0)

	for rows.Next() {
		var (
			id                        string
			name                      string
			monthlyContributionAmount float64
			dueDay                    int
			supportAmount             float64
			penaltyPercentage         float64
			status                    string
		)

		if err := rows.Scan(
			&id,
			&name,
			&monthlyContributionAmount,
			&dueDay,
			&supportAmount,
			&penaltyPercentage,
			&status,
		); err != nil {
			return nil, fmt.Errorf(
				"scan member edir: %w",
				err,
			)
		}

		edirs = append(edirs, map[string]interface{}{
			"id":                          id,
			"name":                        name,
			"monthly_contribution_amount": monthlyContributionAmount,
			"due_ethiopian_day":           dueDay,
			"support_amount":              supportAmount,
			"penalty_percentage":          penaltyPercentage,
			"status":                      status,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate member edirs: %w",
			err,
		)
	}

	return edirs, nil
}
