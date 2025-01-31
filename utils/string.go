package utils

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"strings"
)

func Pluralize(s string, count int64) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", s)
	}

	// hashes, batches
	if strings.HasSuffix(strings.ToLower(s), "h") {
		s += "e"
	}

	return fmt.Sprintf("%s %ss", humanize.Comma(count), s)
}

func PrintFormattedTitle(title string) {
	color.HiCyan(title)
	fmt.Println(strings.Repeat("=", len(title)))
}
