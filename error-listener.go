package sqlparser

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type ErrorListener struct {
	*antlr.DefaultErrorListener
}

var _ antlr.ErrorListener = (*ErrorListener)(nil)

func NewErrorListener() *ErrorListener {
	scanner := new(ErrorListener)
	scanner.DefaultErrorListener = antlr.NewDefaultErrorListener()
	return scanner
}

func (el *ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	str := strings.Builder{}
	str.WriteString("line -- ")
	str.WriteString(strconv.Itoa(line))
	str.WriteString(":")
	str.WriteString(strconv.Itoa(column))
	str.WriteString("\t")
	str.WriteString(msg)
	panic(str.String())
}

func (el *ErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	// TODO:
}
func (el *ErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
	// TODO:
}
func (el *ErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs *antlr.ATNConfigSet) {
	// TODO:
}
