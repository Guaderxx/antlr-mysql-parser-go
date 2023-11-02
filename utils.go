package sqlparser

import (
	"os"
	"strings"
)

func FileExists(name string) bool {
	stats, err := os.Stat(name)
	return err == nil && !stats.IsDir()
}

func WithTrimQuote(str string) string {
	return strings.Trim(str, "'\"`")
}

func WithReplacer(str string, oldnew ...string) string {
	return strings.NewReplacer(oldnew...).Replace(str)
}

func WithTrimBracket(str string) string {
	return strings.Trim(str, "([{}])")
}
