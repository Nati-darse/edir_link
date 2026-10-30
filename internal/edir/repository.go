package edir

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Edir struct {
	ID                        string
	Name                      string
	OrganizerID               string
	MonthlyContributionAmount float64
	DueEthiopianDay           int
	SupportAmount             float64
	PenaltyPercentage         float64
	Status                    string
	RejectionReason           *string
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	name string,
	organizerID string,
	monthlyContributionAmount float64,
	dueDay int,
	supportAmount float64,
	penaltyPercentage float64,
) (*Edir, error) {

	edir := &Edir{}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO edirs (
			name,
			organizer_id,
			monthly_contribution_amount,
			due_ethiopian_day,
			support_amount,
			penalty_percentage,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')
		RETURNING
			id,
			name,
			organizer_id,
			monthly_contribution_amount,
			due_ethiopian_day,
			support_amount,
			penalty_percentage,
			status,
			rejection_reason
		`,
		name,
		organizerID,
		monthlyContributionAmount,
		dueDay,
		supportAmount,
		penaltyPercentage,
	).Scan(
		&edir.ID,
		&edir.Name,
		&edir.OrganizerID,
		&edir.MonthlyContributionAmount,
		&edir.DueEthiopianDay,
		&edir.SupportAmount,
		&edir.PenaltyPercentage,
		&edir.Status,
		&edir.RejectionReason,
	)

	if err != nil {
		return nil, fmt.Errorf("create edir: %w", err)
	}

	return edir, nil
}

func (r *Repository) FindPending(
	ctx context.Context,
) ([]Edir, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
			name,
			organizer_id,
			monthly_contribution_amount,
			due_ethiopian_day,
			support_amount,
			penalty_percentage,
			status,
			rejection_reason
		FROM edirs
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		`,
	)

	if err != nil {
		return nil, fmt.Errorf("find pending edirs: %w", err)
	}

	defer rows.Close()

	var edirs []Edir

	for rows.Next() {
		var edir Edir

		err := rows.Scan(
			&edir.ID,
			&edir.Name,
			&edir.OrganizerID,
			&edir.MonthlyContributionAmount,
			&edir.DueEthiopianDay,
			&edir.SupportAmount,
			&edir.PenaltyPercentage,
			&edir.Status,
			&edir.RejectionReason,
		)

		if err != nil {
			return nil, fmt.Errorf("scan pending edir: %w", err)
		}

		edirs = append(edirs, edir)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending edirs: %w", err)
	}

	return edirs, nil
}

func (r *Repository) Approve(
	ctx context.Context,
	edirID string,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE edirs
		SET
			status = 'ACTIVE',
			rejection_reason = NULL,
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'PENDING'
		`,
		edirID,
	)

	if err != nil {
		return fmt.Errorf("approve edir: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending edir not found")
	}

	return nil
}

func (r *Repository) Reject(
	ctx context.Context,
	edirID string,
	reason string,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE edirs
		SET
			status = 'REJECTED',
			rejection_reason = $2,
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'PENDING'
		`,
		edirID,
		reason,
	)

	if err != nil {
		return fmt.Errorf("reject edir: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending edir not found")
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	edirID string,
) (*Edir, error) {

	edir := &Edir{}

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			organizer_id,
			monthly_contribution_amount,
			due_ethiopian_day,
			support_amount,
			penalty_percentage,
			status,
			rejection_reason
		FROM edirs
		WHERE id = $1
		`,
		edirID,
	).Scan(
		&edir.ID,
		&edir.Name,
		&edir.OrganizerID,
		&edir.MonthlyContributionAmount,
		&edir.DueEthiopianDay,
		&edir.SupportAmount,
		&edir.PenaltyPercentage,
		&edir.Status,
		&edir.RejectionReason,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"find edir by id: %w",
			err,
		)
	}

	return edir, nil
}
