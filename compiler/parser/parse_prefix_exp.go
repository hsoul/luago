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

// var 声明定义变量，params指定函数传递参数
/*
作者：airtrack
链接：https://www.zhihu.com/question/26548261/answer/33818016

提取左因子和消除左递归的方法请参考编译原理的书。Lua中这段BNF可以这么消除：prefixexp ::= var | functioncall | ‘(’ exp ‘)’
functioncall ::=  prefixexp args | prefixexp ‘:’ Name args

==>

prefixexp ::= '(' exp ')' | Name | prefixexp '[' exp ']' | prefixexp '.' Name
            | prefixexp args | prefixexp ':' Name args

==>

prefixexp ::= '(' exp ')' prefixexp_tail | Name prefixexp_tail
prefixexp_tail ::= '[' exp ']' prefixexp_tail |
					'.' Name prefixexp_tail |
					args prefixexp_tail |
					':' Name args prefixexp_tail |
					epsilon -- 空集

PrefixExp 在 Lua 语言中是指“前缀表达式”。
Lua 的语法设计上，PrefixExp 可以是一个变量名、一个函数调用，或者一个表的字段访问。
这种设计是 Lua 语法的一部分，用来识别和解析表达式的起始部分。

Lua 的 PrefixExp 设计允许以下几种形式：

1. 变量 (例如 x)
2. 函数调用 (例如 f() 或 obj:method())
3. 表访问 (例如 t["key"] 或 t.key)
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
			TOKEN_SEP_LPAREN, // (
			TOKEN_SEP_LCURLY, // {
			TOKEN_STRING:     // string literal（字面值） prefixexp args
			exp = _finishFuncCallExp(lexer, exp) // a:b(); print "hello"; print("hello"); print {1, 2}
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
	if lexer.LookAhead() == TOKEN_SEP_COLON { // :
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
