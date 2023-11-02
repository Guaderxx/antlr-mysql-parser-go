package sqlparser

import (
	"github.com/antlr4-go/antlr/v4"
)

func From(name string) ([]*Table, error) {
	var input antlr.CharStream
	var err error
	if FileExists(name) {
		input, err = antlr.NewFileStream(name)
	} else {
		input = antlr.NewInputStream(name)
	}
	if err != nil {
		return nil, err
	}
	lexer := NewMySqlLexer(input)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)
	p := NewMySqlParser(tokens)

	v := new(Visitor)
	var res []*Table

	cts := p.Root().Accept(v)
	if cts == nil {
		return res, nil
	}

	if tmp, ok := cts.([]*CreateTable); ok {
		for _, ct := range tmp {
			res = append(res, ct.Convert())
		}
	}
	return res, nil
}
