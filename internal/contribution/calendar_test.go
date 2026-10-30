package contribution

import "testing"

func TestIsContributionMonth(t *testing.T) {
	for month := 1; month <= 12; month++ {
		if !IsContributionMonth(month) {
			t.Errorf("month %d should allow contributions", month)
		}
	}

	for _, month := range []int{0, 13, 14} {
		if IsContributionMonth(month) {
			t.Errorf("month %d should not allow contributions", month)
		}
	}
}

func TestFirstContributionMonth(t *testing.T) {
	tests := []struct {
		name         string
		currentYear  int
		joinYear     int
		joinMonth    int
		wantStart    int
		wantEligible bool
		wantError    bool
	}{
		{
			name:         "joined this year",
			currentYear:  2019,
			joinYear:     2019,
			joinMonth:    2,
			wantStart:    2,
			wantEligible: true,
		},
		{
			name:         "joined in previous year",
			currentYear:  2019,
			joinYear:     2018,
			joinMonth:    9,
			wantStart:    1,
			wantEligible: true,
		},
		{
			name:        "joined in future year",
			currentYear: 2019,
			joinYear:    2020,
			joinMonth:   1,
		},
		{
			name:        "joined during Pagume this year",
			currentYear: 2019,
			joinYear:    2019,
			joinMonth:   13,
		},
		{
			name:         "joined during previous year's Pagume",
			currentYear:  2019,
			joinYear:     2018,
			joinMonth:    13,
			wantStart:    1,
			wantEligible: true,
		},
		{
			name:        "invalid joining month",
			currentYear: 2019,
			joinYear:    2019,
			joinMonth:   14,
			wantError:   true,
		},
		{
			name:        "invalid Ethiopian year",
			currentYear: 0,
			joinYear:    2019,
			joinMonth:   1,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEligible, err := firstContributionMonth(
				tt.currentYear,
				tt.joinYear,
				tt.joinMonth,
			)

			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}

			if gotStart != tt.wantStart || gotEligible != tt.wantEligible {
				t.Errorf(
					"got start=%d eligible=%v; want start=%d eligible=%v",
					gotStart,
					gotEligible,
					tt.wantStart,
					tt.wantEligible,
				)
			}
		})
	}
}
