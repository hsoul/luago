package parser

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"
)

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args

/*
prefixexp ::=

	Name |
	prefixexp ‘[’ exp ‘]’ |
	prefixexp ‘.’ Name |
	prefixexp [‘:’ Name] args
	‘(’ exp ‘)’ |
*/
func parsePrefixExp(lexer *Lexer) Exp {
	var exp Exp
	if lexer.LookAhead() == TOKEN_IDENTIFIER {
		line, name := lexer.NextIdentifier() // Name
		exp = &NameExp{Line: line, Name: name}
	} else { // ‘(’ exp ‘)’
		exp = parseParensExp(lexer)
	}
	return _finishPrefixExp(lexer, exp)
}

func parseParensExp(lexer *Lexer) Exp { // "Parens"是"Parentheses"的缩写形式，表示圆括号。
	lexer.NextTokenOfKind(TOKEN_SEP_LPAREN) // (
	exp := parseExp(lexer)                  // exp
	lexer.NextTokenOfKind(TOKEN_SEP_RPAREN) // )

	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp: // 由于圆括号会改变vararg和函数调用表达式的语义（详见第8章），所以需要保留这两种语句的圆括号。对于var表达式，也需要保留圆括号，否则前面介绍过的_checkVar（）函数就会出现问题。其余表达式两侧的圆括号则完全没必要留在AST里。
		return &ParensExp{Exp: exp}
	}

	// no need to keep parens
	return exp
}

func _finishPrefixExp(lexer *Lexer, exp Exp) Exp {
	for {
		switch lexer.LookAhead() {
		case TOKEN_SEP_LBRACK: // prefixexp ‘[’ exp ‘]’
			lexer.NextToken()                       // ‘[’
			keyExp := parseExp(lexer)               // exp
			lexer.NextTokenOfKind(TOKEN_SEP_RBRACK) // ‘]’
			exp = &TableAccessExp{LastLine: lexer.Line(), PrefixExp: exp, KeyExp: keyExp}
		case TOKEN_SEP_DOT: // prefixexp ‘.’ Name
			lexer.NextToken()                    // ‘.’
			line, name := lexer.NextIdentifier() // Name
			keyExp := &StringExp{Line: line, Str: name}
			exp = &TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: keyExp}
		case TOKEN_SEP_COLON, // prefixexp ‘:’ Name args
			TOKEN_SEP_LPAREN, TOKEN_SEP_LCURLY, TOKEN_STRING: // prefixexp args
			exp = _finishFuncCallExp(lexer, exp)
		default:
			return exp
		}
	}
	return exp
}

// functioncall ::=  prefixexp args | prefixexp ‘:’ Name args
func _finishFuncCallExp(lexer *Lexer, prefixExp Exp) *FuncCallExp {
	nameExp := _parseNameExp(lexer)
	line := lexer.Line() // todo
	args := _parseArgs(lexer)
	lastLine := lexer.Line()
	return &FuncCallExp{Line: line, LastLine: lastLine, PrefixExp: prefixExp, NameExp: nameExp, Args: args}
}

func _parseNameExp(lexer *Lexer) *StringExp {
	if lexer.LookAhead() == TOKEN_SEP_COLON {
		lexer.NextToken()
		line, name := lexer.NextIdentifier()
		return &StringExp{Line: line, Str: name}
	}
	return nil
}

// args ::=  ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func _parseArgs(lexer *Lexer) (args []Exp) {
	switch lexer.LookAhead() {
	case TOKEN_SEP_LPAREN: // ‘(’ [explist] ‘)’
		lexer.NextToken() // TOKEN_SEP_LPAREN
		if lexer.LookAhead() != TOKEN_SEP_RPAREN {
			args = parseExpList(lexer)
		}
		lexer.NextTokenOfKind(TOKEN_SEP_RPAREN)
	case TOKEN_SEP_LCURLY: // ‘{’ [fieldlist] ‘}’
		args = []Exp{parseTableConstructorExp(lexer)}
	default: // LiteralString
		line, str := lexer.NextTokenOfKind(TOKEN_STRING)
		args = []Exp{&StringExp{Line: line, Str: str}}
	}
	return
}
