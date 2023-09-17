package internal

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/avamsi/ergo/assert"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	anyUpperishLower = regexp.MustCompile("(.)([A-Z][0-9]*)([a-z])")
	lowerishUpper    = regexp.MustCompile("([a-z][0-9]*)([A-Z])")
	invalids         = regexp.MustCompile("[^a-zA-Z0-9]+")
)

// NormalizeToKebabCase normalizes the input string to ASCII kebab-case.
// It tries to convert non-ASCII runes in the input string to ASCII by
// decomposing and then dropping all non-ASCII runes (and so is lossy).
// It supports camelCase, PascalCase, snake_case, and SCREAMING_SNAKE_CASE --
// anything else (including digits mixed in) working is a happy accident.
func NormalizeToKebabCase(s string) string {
	// Decompose and remove all non-spacing marks.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)))
	s, _, err := transform.String(t, s)
	assert.Nil(err)
	s = anyUpperishLower.ReplaceAllString(s, "${1}-${2}${3}")
	s = lowerishUpper.ReplaceAllString(s, "${1}-${2}")
	s = invalids.ReplaceAllString(s, "-")
	return strings.ToLower(strings.Trim(s, "-"))
}
