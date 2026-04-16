package task

import "time"

// ShouldGenerateOnDate определяет, нужно ли создать экземпляр задачи на указанную дату.
// baseDate — ScheduledAt исходного шаблона.
func ShouldGenerateOnDate(r *Recurrence, date time.Time, baseDate time.Time) bool {
	if r == nil {
		return false
	}
	date = date.Truncate(24 * time.Hour)
	baseDate = baseDate.Truncate(24 * time.Hour)

	switch r.Type {
	case RecurrenceDaily:
		if date.Before(baseDate) {
			return false
		}
		daysDiff := int(date.Sub(baseDate).Hours() / 24)
		return daysDiff%r.Interval == 0
	case RecurrenceMonthly:
		if date.Before(baseDate) {
			return false
		}
		targetDay := r.DayOfMonth
		lastDay := daysInMonth(date.Year(), date.Month())
		if targetDay > lastDay {
			targetDay = lastDay
		}
		return date.Day() == targetDay
	case RecurrenceSpecific:
		dateStr := date.Format("2006-01-02")
		for _, d := range r.Dates {
			if d == dateStr {
				return true
			}
		}
		return false
	case RecurrenceParity:
		if date.Before(baseDate) {
			return false
		}
		day := date.Day()
		if r.Parity == "even" {
			return day%2 == 0
		}
		return day%2 == 1
	default:
		return false
	}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}