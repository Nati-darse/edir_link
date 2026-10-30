package ethiopiancalendar

import (
	"testing"
	"time"
)

func TestFromGregorianKnownDates(t *testing.T) {
	cases := []struct {
		y    int
		m    time.Month
		d    int
		want Date
	}{
		{2026, time.October, 14, Date{2019, 2, 4}}, // date stated by the product owner
		{2026, time.October, 10, Date{2019, 1, 30}},
		{2026, time.October, 11, Date{2019, 2, 1}},
		{2026, time.September, 11, Date{2019, 1, 1}},  // new year after a non-leap year
		{2007, time.September, 12, Date{2000, 1, 1}},  // Millennium
		{2007, time.September, 11, Date{1999, 13, 6}}, // Pagume of a leap year
		{2023, time.September, 12, Date{2016, 1, 1}},
		{2023, time.September, 11, Date{2015, 13, 6}},
		{2022, time.September, 11, Date{2015, 1, 1}},
		{2022, time.September, 10, Date{2014, 13, 5}}, // Pagume of a non-leap year
		{2022, time.September, 6, Date{2014, 13, 1}},
		{2022, time.June, 15, Date{2014, 10, 8}},
		{2022, time.March, 1, Date{2014, 6, 22}},
		{2021, time.December, 10, Date{2014, 4, 1}},
		{2024, time.April, 19, Date{2016, 8, 11}}, // Gregorian leap year
		{2024, time.February, 29, Date{2016, 6, 21}},
		{2000, time.January, 1, Date{1992, 4, 22}},
	}
	for _, c := range cases {
		got, err := FromGregorian(c.y, c.m, c.d)
		if err != nil {
			t.Fatalf("%d-%02d-%02d: %v", c.y, c.m, c.d, err)
		}
		if got != c.want {
			t.Errorf("%d-%02d-%02d: got %v, want %v", c.y, c.m, c.d, got, c.want)
		}
	}
}

func TestRoundTripAndContinuity(t *testing.T) {
	day := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2100, 12, 31, 0, 0, 0, 0, time.UTC)
	prev, err := FromGregorian(day.Year(), day.Month(), day.Day())
	if err != nil {
		t.Fatal(err)
	}
	for day = day.AddDate(0, 0, 1); !day.After(end); day = day.AddDate(0, 0, 1) {
		cur, err := FromGregorian(day.Year(), day.Month(), day.Day())
		if err != nil {
			t.Fatal(err)
		}
		if err := cur.Validate(); err != nil {
			t.Fatalf("%v produced invalid %v: %v", day, cur, err)
		}
		back, err := cur.ToGregorian()
		if err != nil || !back.Equal(day) {
			t.Fatalf("round trip failed: %v -> %v -> %v (%v)", day, cur, back, err)
		}
		// Each Gregorian day must advance the Ethiopian date by exactly one day.
		if !prev.Before(cur) || cur.daysSinceEpoch()-prev.daysSinceEpoch() != 1 {
			t.Fatalf("not consecutive: %v then %v", prev, cur)
		}
		prev = cur
	}
}

func TestPagumeLength(t *testing.T) {
	cases := []struct {
		year, days int
	}{{2018, 5}, {2019, 6}, {2020, 5}, {2015, 6}}
	for _, c := range cases {
		got, err := DaysInMonth(c.year, 13)
		if err != nil || got != c.days {
			t.Errorf("Pagume %d: got %d (%v), want %d", c.year, got, err, c.days)
		}
	}
	if err := (Date{2018, 13, 6}).Validate(); err == nil {
		t.Error("Pagume 6 of a non-leap year must be invalid")
	}
	if err := (Date{2019, 13, 6}).Validate(); err != nil {
		t.Errorf("Pagume 6 of a leap year must be valid: %v", err)
	}
}

func TestValidateRejectsBadDates(t *testing.T) {
	bad := []Date{
		{0, 1, 1}, {2019, 0, 1}, {2019, 14, 1}, {2019, 1, 0}, {2019, 1, 31}, {2019, 13, 7},
	}
	for _, d := range bad {
		if d.Validate() == nil {
			t.Errorf("%v should be invalid", d)
		}
	}
}

func TestServiceUsesAddisAbabaTimezone(t *testing.T) {
	// 21:00 UTC is midnight in Addis Ababa (UTC+3).
	cases := []struct {
		at   time.Time
		want Date
	}{
		{time.Date(2026, 10, 13, 20, 59, 59, 0, time.UTC), Date{2019, 2, 3}},
		{time.Date(2026, 10, 13, 21, 0, 0, 0, time.UTC), Date{2019, 2, 4}},
		{time.Date(2026, 10, 14, 20, 59, 59, 0, time.UTC), Date{2019, 2, 4}},
		{time.Date(2026, 10, 14, 21, 0, 0, 0, time.UTC), Date{2019, 2, 5}},
	}
	for _, c := range cases {
		at := c.at
		svc, err := NewWithClock(func() time.Time { return at })
		if err != nil {
			t.Fatal(err)
		}
		if got := svc.Today(); got != c.want {
			t.Errorf("at %v: got %v, want %v", c.at, got, c.want)
		}
	}
}

func TestServiceAdvancesWithClock(t *testing.T) {
	now := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	svc, err := NewWithClock(func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if got := svc.Today(); got != (Date{2019, 2, 4}) {
		t.Fatalf("got %v", got)
	}
	now = now.Add(24 * time.Hour)
	if got := svc.Today(); got != (Date{2019, 2, 5}) {
		t.Fatalf("clock advance not reflected: %v", got)
	}
}

func TestYearBoundary(t *testing.T) {
	// 2019 is a leap year: Pagume has 6 days, new year is 12 September 2027.
	svc := func(at time.Time) Date {
		s, _ := NewWithClock(func() time.Time { return at })
		return s.Today()
	}
	noon := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 9, 0, 0, 0, time.UTC)
	}
	if got := svc(noon(2027, time.September, 11)); got != (Date{2019, 13, 6}) {
		t.Errorf("got %v", got)
	}
	if got := svc(noon(2027, time.September, 12)); got != (Date{2020, 1, 1}) {
		t.Errorf("got %v", got)
	}
}

func TestDueDateAndOverdue(t *testing.T) {
	today := Date{2019, 2, 10}
	// Due on Tikimt 10: still payable on the due day.
	over, err := IsOverdue(2019, 2, 10, today)
	if err != nil || over {
		t.Errorf("due day itself must not be overdue (%v, %v)", over, err)
	}
	// Overdue from the next day.
	over, err = IsOverdue(2019, 2, 10, Date{2019, 2, 11})
	if err != nil || !over {
		t.Errorf("day after due day must be overdue (%v, %v)", over, err)
	}
	// Previous month obligation is overdue across a month boundary.
	over, _ = IsOverdue(2019, 1, 30, Date{2019, 2, 1})
	if !over {
		t.Error("Meskerem 30 due date must be overdue on Tikimt 1")
	}
	// Previous year obligation is overdue.
	over, _ = IsOverdue(2018, 12, 10, Date{2019, 1, 1})
	if !over {
		t.Error("previous-year obligation must be overdue")
	}
	if _, err := IsOverdue(2019, 13, 1, today); err == nil {
		t.Error("Pagume must not be a contribution month")
	}
}

func TestContributionMonth(t *testing.T) {
	for m := 1; m <= 12; m++ {
		if !IsContributionMonth(m) {
			t.Errorf("month %d should be a contribution month", m)
		}
	}
	if IsContributionMonth(0) || IsContributionMonth(13) {
		t.Error("months 0 and 13 are not contribution months")
	}
}
