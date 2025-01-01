package utils

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
)

func Pluralize(word string, count int64) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, word)
	}

	return fmt.Sprintf("%d %ss", count, word)
}

func PrintFormattedTitle(title string) {
	color.HiCyan(title)
	fmt.Println(strings.Repeat("=", len(title)))
}
