package rails

import "strings"

// Filename implements ActiveStorage::Filename#sanitized for display and indexing.
func Filename(name string) string {
	return strings.Map(func(c rune) rune {
		if strings.ContainsRune("\u202e%$|:;/<>?*\"\t\r\n\\", c) {
			return '-'
		}
		return c
	}, strings.Trim(name, "\x00\t\n\v\f\r "))
}
