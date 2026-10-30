package contribution

import "fmt"

// IsContributionMonth returns false for Pagume (month 13).
// Contribution obligations are generated only for Ethiopian months 1 through 12.
func IsContributionMonth(month int) bool {
	return month >= 1 && month <= 12
}

func firstContributionMonth(currentYear, joinYear, joinMonth int) (int, bool, error) {
	if currentYear < 1 || joinYear < 1 {
		return 0, false, fmt.Errorf("Ethiopian years must be positive")
	}
	if joinMonth < 1 || joinMonth > 13 {
		return 0, false, fmt.Errorf("Ethiopian joining month must be between 1 and 13")
	}
	if joinYear > currentYear || (joinYear == currentYear && joinMonth == 13) {
		return 0, false, nil
	}
	if joinYear < currentYear {
		return 1, true, nil
	}
	return joinMonth, true, nil
}
