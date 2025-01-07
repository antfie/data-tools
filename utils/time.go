package utils

import (
	"fmt"
	"strings"
	"time"
)

func FormatDuration(duration time.Duration) string {
	str := duration.String()

	if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		remainingHours := duration.Hours() - float64(days*24)
		str = fmt.Sprintf("%dd %dh%s", days, int(remainingHours), strings.Split(str, "h")[1])
	}

	// Deal with time units less than a minute, e.g. seconds, ms,μs,ns
	if strings.Contains(str, "ms") || !strings.Contains(str, "m") {
		return str
	}

	// Reduce precision of times for readability
	str = strings.Split(str, "m")[0]

	// Formatting
	str = strings.Replace(str, "h", "h ", 1) + "m"

	return str
}
