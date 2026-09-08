package dateUtils

import (
	"fmt"
	"time"
)

const (
	dateFormat   = "Jan 2, 2006"
	hoursPerDay  = 24
	hoursPerWeek = hoursPerDay * 7
)

func HumanTimeSince(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return date
	}

	timeSince := time.Since(t)

	if timeSince > hoursPerWeek*time.Hour {
		return t.Format(dateFormat)
	}

	units, unitName := GetTimeUnit(t)

	if units > 1 { // Every number greater that 1 uses the plural form of the unit
		unitName += "s"
	}

	return fmt.Sprintf("%d %s ago", int(units), unitName)
}

func GetTimeUnit(t time.Time) (float64, string) {
	timeSince := time.Since(t)

	if timeSince < time.Minute {
		return timeSince.Seconds(), "second"
	}

	if timeSince < time.Hour {
		return timeSince.Minutes(), "minute"
	}

	hours := timeSince.Hours()

	if timeSince < hoursPerDay*time.Hour {
		return hours, "hour"
	}

	return hours / hoursPerDay, "day"
}
