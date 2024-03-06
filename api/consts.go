package api

const LUA_MINSTACK = 20
const LUAI_MAXSTACK = 1000000
const LUA_REGISTRYINDEX = -LUAI_MAXSTACK - 1000
const LUA_RIDX_MAINTHREAD int64 = 1
const LUA_RIDX_GLOBALS int64 = 2
const LUA_MULTRET = -1

const (
	LUA_MAXINTEGER = 1<<63 - 1
	LUA_MININTEGER = -1 << 63
)

// 在语言层面，Lua一共支持8种数据类型，分别是nil、布尔（boolean）、数字（number）、字符串（string）、表（table）、函数（function）、线程（thread）和用户数据（userdata）

// LuaType是一个整数类型，每个常量都代表了一个Lua类型
const (
	LUA_TNONE          = iota - 1 // -1
	LUA_TNIL                      // 0
	LUA_TBOOLEAN                  // 1
	LUA_TLIGHTUSERDATA            // 2
	LUA_TNUMBER                   // 3
	LUA_TSTRING                   // 4
	LUA_TTABLE                    // 5
	LUA_TFUNCTION                 // 6
	LUA_TUSERDATA                 // 7
	LUA_TTHREAD                   // 8
)

// 数学运算符
const (
	LUA_OPADD  = iota // +
	LUA_OPSUB         // -
	LUA_OPMUL         // *
	LUA_OPMOD         // %
	LUA_OPPOW         // ^
	LUA_OPDIV         // /
	LUA_OPIDIV        // // 整除
	LUA_OPBAND        // &
	LUA_OPBOR         // |
	LUA_OPBXOR        // ~
	LUA_OPSHL         // <<
	LUA_OPSHR         // >>
	LUA_OPUNM         // - (unary minus)
	LUA_OPBNOT        // ~
)

// 比较运算符
const (
	LUA_OPEQ = iota // ==
	LUA_OPLT        // <
	LUA_OPLE        // <=
)

const (
	LUA_OK = iota
	LUA_YIELD
	LUA_ERRRUN
	LUA_ERRSYNTAX
	LUA_ERRMEM
	LUA_ERRGCMM
	LUA_ERRERR
	LUA_ERRFILE
)
