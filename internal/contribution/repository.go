package contribution

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nati-darse/edirlink-backend/internal/ethiopiancalendar"
)

type Contribution struct {
	ID              string     `json:"id"`
	EdirID          string     `json:"edir_id"`
	MemberID        string     `json:"member_id"`
	EthiopianYear   int        `json:"ethiopian_year"`
	EthiopianMonth  int        `json:"ethiopian_month"`
	DueEthiopianDay int        `json:"due_ethiopian_day"`
	Amount          float64    `json:"amount"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at"`
	PenaltyAmount   float64    `json:"penalty_amount"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Repository struct {
	db       *pgxpool.Pool
	calendar *ethiopiancalendar.Service
}

func NewRepository(db *pgxpool.Pool, calendar *ethiopiancalendar.Service) *Repository {
	return &Repository{db: db, calendar: calendar}
}

const columns = `id, edir_id, member_id, ethiopian_year, ethiopian_month,
	due_ethiopian_day, amount, status, paid_at, penalty_amount, created_at, updated_at`

func scan(row interface{ Scan(...any) error }) (*Contribution, error) {
	c := &Contribution{}
	err := row.Scan(
		&c.ID, &c.EdirID, &c.MemberID,
		&c.EthiopianYear, &c.EthiopianMonth, &c.DueEthiopianDay,
		&c.Amount, &c.Status, &c.PaidAt, &c.PenaltyAmount,
		&c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

// Create accepts Ethiopian calendar values directly; no Gregorian business dates are required.
func (r *Repository) Create(
	ctx context.Context,
	edirID, memberID string,
	year, month, dueDay int,
	amount float64,
) (*Contribution, error) {
	if amount < 0 {
		return nil, fmt.Errorf("contribution amount cannot be negative")
	}
	if year < 1 {
		return nil, fmt.Errorf("Ethiopian year must be positive")
	}
	if !IsContributionMonth(month) {
		return nil, fmt.Errorf("Ethiopian contribution month must be between 1 and 12; Pagume is exempt")
	}
	if dueDay < 1 || dueDay > 30 {
		return nil, fmt.Errorf("Ethiopian due day must be between 1 and 30")
	}

	c, err := scan(r.db.QueryRow(ctx, `
		INSERT INTO contributions
			(edir_id, member_id, ethiopian_year, ethiopian_month,
			 due_ethiopian_day, amount, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')
		RETURNING `+columns,
		edirID, memberID, year, month, dueDay, amount,
	))
	if err != nil {
		return nil, fmt.Errorf("create contribution: %w", err)
	}
	return c, nil
}

func (r *Repository) ListByMember(ctx context.Context, memberID string) ([]*Contribution, error) {
	return r.list(ctx, `SELECT `+columns+` FROM contributions
		WHERE member_id=$1
		ORDER BY ethiopian_year DESC, ethiopian_month DESC`, memberID)
}

func (r *Repository) ListByEdir(ctx context.Context, edirID string) ([]*Contribution, error) {
	return r.list(ctx, `SELECT `+columns+` FROM contributions
		WHERE edir_id=$1
		ORDER BY ethiopian_year DESC, ethiopian_month DESC, created_at DESC`, edirID)
}

func (r *Repository) list(ctx context.Context, query, id string) ([]*Contribution, error) {
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	defer rows.Close()

	items := make([]*Contribution, 0)
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan contribution: %w", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) OrganizerOwnsEdir(ctx context.Context, edirID, organizerID string) (bool, error) {
	var owns bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM edirs WHERE id=$1 AND organizer_id=$2)`,
		edirID, organizerID,
	).Scan(&owns)
	return owns, err
}
