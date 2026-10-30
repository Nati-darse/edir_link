package contribution

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nati-darse/edirlink-backend/internal/ethiopiancalendar"
)

type membershipStart struct {
	memberID string
	year     pgtype.Int4
	month    pgtype.Int2
	day      pgtype.Int2
}

func (r *Repository) GenerateForEdir(ctx context.Context, edirID string) (int, error) {
	today := r.calendar.Today()
	if err := today.Validate(); err != nil {
		return 0, fmt.Errorf("invalid current Ethiopian date: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin contribution generation: %w", err)
	}
	defer tx.Rollback(ctx)

	var amount float64
	var dueDay int
	var status string

	err = tx.QueryRow(ctx, `
SELECT monthly_contribution_amount,
       due_ethiopian_day,
       status
FROM edirs
WHERE id = $1
`, edirID).Scan(&amount, &dueDay, &status)
	if err != nil {
		return 0, fmt.Errorf("load Edir contribution settings: %w", err)
	}

	if status != "ACTIVE" {
		return 0, fmt.Errorf("Edir must be active")
	}
	if amount < 0 || dueDay < 1 || dueDay > 30 {
		return 0, fmt.Errorf("invalid Edir contribution settings")
	}

	rows, err := tx.Query(ctx, `
SELECT member_id,
       joined_ethiopian_year,
       joined_ethiopian_month,
       joined_ethiopian_day
FROM memberships
WHERE edir_id = $1 AND is_active = TRUE
`, edirID)
	if err != nil {
		return 0, fmt.Errorf("load active members: %w", err)
	}

	members := make([]membershipStart, 0)
	for rows.Next() {
		var m membershipStart
		if err := rows.Scan(&m.memberID, &m.year, &m.month, &m.day); err != nil {
			rows.Close()
			return 0, fmt.Errorf("read member Ethiopian joining date: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate active members: %w", err)
	}
	rows.Close()

	// Validate every active member before inserting anything. Legacy memberships
	// with unknown Ethiopian joining dates must be corrected explicitly.
	for _, member := range members {
		if !member.year.Valid || !member.month.Valid || !member.day.Valid {
			return 0, fmt.Errorf(
				"member %s has no Ethiopian joining date; set it before generating contributions",
				member.memberID,
			)
		}
		joined := ethiopiancalendar.Date{
			Year:  int(member.year.Int32),
			Month: int(member.month.Int16),
			Day:   int(member.day.Int16),
		}
		if err := joined.Validate(); err != nil {
			return 0, fmt.Errorf("member %s has an invalid Ethiopian joining date: %w", member.memberID, err)
		}
	}

	// Never generate beyond the current month, and never for Pagume.
	lastMonth := today.Month
	if lastMonth > ethiopiancalendar.LastContributionMonth {
		lastMonth = ethiopiancalendar.LastContributionMonth
	}

	created := 0
	for _, member := range members {
		startMonth, eligible, err := firstContributionMonth(
			today.Year,
			int(member.year.Int32),
			int(member.month.Int16),
		)
		if err != nil {
			return 0, err
		}
		if !eligible {
			continue
		}

		// Existing periods and amounts are preserved by ON CONFLICT DO NOTHING.
		for month := startMonth; month <= lastMonth; month++ {
			tag, err := tx.Exec(ctx, `
INSERT INTO contributions (
edir_id, member_id, ethiopian_year, ethiopian_month,
due_ethiopian_day, amount, status
)
VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')
ON CONFLICT (edir_id, member_id, ethiopian_year, ethiopian_month)
DO NOTHING
`, edirID, member.memberID, today.Year, month, dueDay, amount)
			if err != nil {
				return 0, fmt.Errorf("generate contribution for member %s month %d: %w",
					member.memberID, month, err)
			}
			created += int(tag.RowsAffected())
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit contribution generation: %w", err)
	}
	return created, nil
}
