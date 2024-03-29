package climate

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/avamsi/ergo/assert"
)

func parseBool(s string) bool {
	return assert.Ok(strconv.ParseBool(s))
}

func parseInt64(s string) int64 {
	return assert.Ok(strconv.ParseInt(s, 10, 64))
}

func parseUint64(s string) uint64 {
	return assert.Ok(strconv.ParseUint(s, 10, 64))
}

func parseFloat64(s string) float64 {
	return assert.Ok(strconv.ParseFloat(s, 64))
}

func parseString(s string) string {
	return s
}

type typeParser[T any] func(string) T

func sliceParser[T any](typer typeParser[T]) typeParser[[]T] {
	return func(s string) []T {
		// Plumb through csv.Reader (instead of strings.Split(s, ",") or
		// something similar) to account for quotes etc.
		ss := assert.Ok(csv.NewReader(strings.NewReader(s)).Read())
		var ts []T
		for _, s := range ss {
			if s := strings.TrimSpace(s); s != "" {
				ts = append(ts, typer(s))
			}
		}
		return ts
	}
}
