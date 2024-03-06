package ast

// 语句（Statement）是最基本的执行单位，表达式（Expression）则是构成语句的要素之一。
// 语句和表达式的主要区别在于：语句只能执行不能用于求值，而表达式只能用于求值不能单独执行。
// 语句和表达式也并非泾渭分明，比如在Lua里，函数调用既可以是表达式，也可以是语句。

// Lua语句大致可以分为控制语句、声明和赋值语句、以及其他语句三类。
// 声明和赋值语句用于声明局部变量、给变量赋值或者往表里写入值，包括局部变量声明语句（见15.3.6节）、赋值语句（见15.3.7节）、局部函数定义语句（见15.3.9节）和非局部函数定义语句（见15.3.8节）。
// 控制语句用于改变执行流程，包括while和repeat语句（见15.3.2节）、if语句（见15.3.3节）、for循环语句（见15.3.4和15.3.5节）、break语句及label和goto语句。
// 其他语句包括空语句、do语句和函数调用语句。

// chunk ::= block
// type Chunk *Block

// block ::= {stat} [retstat]
// retstat ::= return [explist] [‘;’]
// explist ::= exp {‘,’ exp}
type Block struct {
	LastLine int
	Stats    []Stat
	RetExps  []Exp
}
