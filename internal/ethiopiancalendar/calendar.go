package ethiopiancalendar

import (
	"fmt"
	"time"
	_ "time/tzdata"
)

// LocationName is the business timezone of the platform.
const LocationName = "Africa/Addis_Ababa"

const (
	// Julian Day Number of 1 Meskerem, year 1 (29 August 8 CE, Julian).
	ethiopianEpochJDN = 1724221
	// Julian Day Number of 1970-01-01.
	unixEpochJDN = 2440588

	// PagumeMonth is the 13th month. It is never a contribution month.
	PagumeMonth = 13
	// LastContributionMonth is Nehase.
	LastContributionMonth = 12
)

var monthNames = [...]string{
	"", "Meskerem", "Tikimt", "Hidar", "Tahsas", "Tir", "Yekatit",
	"Megabit", "Miazia", "Genbot", "Sene", "Hamle", "Nehase", "Pagume",
}

// Date is an Ethiopian calendar date.
type Date struct {
	Year  int `json:"ethiopian_year"`
	Month int `json:"ethiopian_month"`
	Day   int `json:"ethiopian_day"`
}

// IsLeapYear reports whether the Ethiopian year has a 6-day Pagume.
func IsLeapYear(year int) bool {
	return year%4 == 3
}

// DaysInMonth returns the number of days in the given Ethiopian month.
func DaysInMonth(year, month int) (int, error) {
	switch {
	case year < 1:
		return 0, fmt.Errorf("Ethiopian year must be positive")
	case month < 1 || month > PagumeMonth:
		return 0, fmt.Errorf("Ethiopian month must be between 1 and 13")
	case month < PagumeMonth:
		return 30, nil
	case IsLeapYear(year):
		return 6, nil
	default:
		return 5, nil
	}
}

// IsContributionMonth reports whether contributions exist for this month
// (months 1-12; Pagume is excluded).
func IsContributionMonth(month int) bool {
	return month >= 1 && month <= LastContributionMonth
}

// MonthName returns the transliterated month name, or "" if out of range.
func MonthName(month int) string {
	if month < 1 || month > PagumeMonth {
		return ""
	}
	return monthNames[month]
}

// Validate checks the date, including Pagume length for leap/non-leap years.
func (d Date) Validate() error {
	days, err := DaysInMonth(d.Year, d.Month)
	if err != nil {
		return err
	}
	if d.Day < 1 || d.Day > days {
		return fmt.Errorf(
			"Ethiopian day must be between 1 and %d for %d/%d",
			days, d.Year, d.Month,
		)
	}
	return nil
}

// Compare returns -1, 0 or 1 when d is before, equal to or after other.
func (d Date) Compare(other Date) int {
	switch {
	case d.Year != other.Year:
		return cmp(d.Year, other.Year)
	case d.Month != other.Month:
		return cmp(d.Month, other.Month)
	default:
		return cmp(d.Day, other.Day)
	}
}

func (d Date) Before(other Date) bool { return d.Compare(other) < 0 }
func (d Date) After(other Date) bool  { return d.Compare(other) > 0 }

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// MonthName returns the name of the date's month.
func (d Date) MonthName() string { return MonthName(d.Month) }

// DueDate builds the due date of a contribution period.
func DueDate(year, month, dueDay int) (Date, error) {
	if !IsContributionMonth(month) {
		return Date{}, fmt.Errorf("month %d is not a contribution month", month)
	}
	d := Date{Year: year, Month: month, Day: dueDay}
	if err := d.Validate(); err != nil {
		return Date{}, err
	}
	return d, nil
}

// IsOverdue reports whether an unpaid obligation is overdue on `today`.
// The obligation stays payable through its due day; it becomes overdue on the
// following Ethiopian day.
func IsOverdue(year, month, dueDay int, today Date) (bool, error) {
	due, err := DueDate(year, month, dueDay)
	if err != nil {
		return false, err
	}
	return today.After(due), nil
}

func yearStartDays(year int) int {
	return 365*(year-1) + year/4
}

func (d Date) daysSinceEpoch() int {
	return yearStartDays(d.Year) + 30*(d.Month-1) + (d.Day - 1)
}

func fromDaysSinceEpoch(n int) Date {
	year := 4*n/1461 + 1
	for yearStartDays(year+1) <= n {
		year++
	}
	for yearStartDays(year) > n {
		year--
	}
	dayOfYear := n - yearStartDays(year)
	return Date{Year: year, Month: dayOfYear/30 + 1, Day: dayOfYear%30 + 1}
}

// FromGregorian converts a Gregorian calendar date to an Ethiopian date.
func FromGregorian(year int, month time.Month, day int) (Date, error) {
	g := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if g.Year() != year || g.Month() != month || g.Day() != day {
		return Date{}, fmt.Errorf("invalid Gregorian date %04d-%02d-%02d", year, int(month), day)
	}
	unixDay := int(g.Unix() / 86400)
	n := unixDay + unixEpochJDN - ethiopianEpochJDN
	if n < 0 {
		return Date{}, fmt.Errorf("date is before the Ethiopian epoch")
	}
	return fromDaysSinceEpoch(n), nil
}

// ToGregorian converts the Ethiopian date to a Gregorian date (UTC midnight).
func (d Date) ToGregorian() (time.Time, error) {
	if err := d.Validate(); err != nil {
		return time.Time{}, err
	}
	unixDay := d.daysSinceEpoch() + ethiopianEpochJDN - unixEpochJDN
	return time.Unix(int64(unixDay)*86400, 0).UTC(), nil
}

// ---- service --------------------------------------------------------------

// Service provides the current Ethiopian date from a clock. Inject a fixed
// clock in tests.
type Service struct {
	now func() time.Time
	loc *time.Location
}

// New returns a Service using the real system clock.
func New() (*Service, error) {
	return NewWithClock(time.Now)
}

// NewWithClock returns a Service using the supplied clock (for tests).
func NewWithClock(now func() time.Time) (*Service, error) {
	loc, err := time.LoadLocation(LocationName)
	if err != nil {
		return nil, fmt.Errorf("load Ethiopian business timezone: %w", err)
	}
	return &Service{now: now, loc: loc}, nil
}

// Today returns today's Ethiopian date in the Addis Ababa timezone.
func (s *Service) Today() Date {
	return s.DateOf(s.now())
}

// DateOf converts an instant (for example a payment timestamp) to the
// Ethiopian business date in the Addis Ababa timezone.
func (s *Service) DateOf(t time.Time) Date {
	local := t.In(s.loc)
	d, err := FromGregorian(local.Year(), local.Month(), local.Day())
	if err != nil {
		// Unreachable for any realistic clock value (dates after 8 CE).
		panic(fmt.Sprintf("ethiopiancalendar: %v", err))
	}
	return d
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
