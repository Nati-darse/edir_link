package ethiopiancalendar

import (
	"encoding/json"
	"net/http"
)

// TodayHandler serves the current Ethiopian date so clients never have to
// compute it themselves. GET /api/calendar/today
func (s *Service) TodayHandler(w http.ResponseWriter, r *http.Request) {
	today := s.Today()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ethiopian_year":  today.Year,
		"ethiopian_month": today.Month,
		"ethiopian_day":   today.Day,
		"month_name":      today.MonthName(),
		"timezone":        LocationName,
	})
}
