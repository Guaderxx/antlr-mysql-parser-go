package sqlparser

import "github.com/antlr4-go/antlr/v4"

type Parser struct {
	*BaseMySqlParserListener
}

var _ MySqlParserListener = (*Parser)(nil)

func (p *Parser) EnterEveryRule(ctx antlr.ParserRuleContext) {
	// TODO:
}

func (p *Parser) ExitEveryRule(ctx antlr.ParserRuleContext) {
	// TODO:
}
