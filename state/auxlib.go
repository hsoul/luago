package state

import (
	"fmt"
	"io/ioutil"

	. "luago/api"
	"luago/stdlib"
)

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_error
func (l *luaState) Error2(fmt string, a ...interface{}) int {
	l.PushFString(fmt, a...) // todo
	return l.Error()
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_argerror
func (l *luaState) ArgError(arg int, extraMsg string) int {
	// bad argument #arg to 'funcname' (extramsg)
	return l.Error2("bad argument #%d (%s)", arg, extraMsg) // todo
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkstack
func (l *luaState) CheckStack2(sz int, msg string) {
	if !l.CheckStack(sz) {
		if msg != "" {
			l.Error2("stack overflow (%s)", msg)
		} else {
			l.Error2("stack overflow")
		}
	}
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_argcheck
func (l *luaState) ArgCheck(cond bool, arg int, extraMsg string) {
	if !cond {
		l.ArgError(arg, extraMsg)
	}
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkany
func (l *luaState) CheckAny(arg int) {
	if l.Type(arg) == LUA_TNONE {
		l.ArgError(arg, "value expected")
	}
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checktype
func (l *luaState) CheckType(arg int, t LuaType) {
	if l.Type(arg) != t {
		l.tagError(arg, t)
	}
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkinteger
func (l *luaState) CheckInteger(arg int) int64 {
	i, ok := l.ToIntegerX(arg)
	if !ok {
		l.intError(arg)
	}
	return i
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checknumber
func (l *luaState) CheckNumber(arg int) float64 {
	f, ok := l.ToNumberX(arg)
	if !ok {
		l.tagError(arg, LUA_TNUMBER)
	}
	return f
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_checkstring
// http://www.lua.org/manual/5.3/manual.html#luaL_checklstring
func (l *luaState) CheckString(arg int) string {
	s, ok := l.ToStringX(arg)
	if !ok {
		l.tagError(arg, LUA_TSTRING)
	}
	return s
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optinteger
func (l *luaState) OptInteger(arg int, def int64) int64 {
	if l.IsNoneOrNil(arg) {
		return def
	}
	return l.CheckInteger(arg)
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optnumber
func (l *luaState) OptNumber(arg int, def float64) float64 {
	if l.IsNoneOrNil(arg) {
		return def
	}
	return l.CheckNumber(arg)
}

// [-0, +0, v]
// http://www.lua.org/manual/5.3/manual.html#luaL_optstring
func (l *luaState) OptString(arg int, def string) string {
	if l.IsNoneOrNil(arg) {
		return def
	}
	return l.CheckString(arg)
}

// [-0, +?, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_dofile
func (l *luaState) DoFile(filename string) bool {
	return l.LoadFile(filename) != LUA_OK ||
		l.PCall(0, LUA_MULTRET, 0) != LUA_OK
}

// [-0, +?, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_dostring
func (l *luaState) DoString(str string) bool {
	return l.LoadString(str) != LUA_OK ||
		l.PCall(0, LUA_MULTRET, 0) != LUA_OK
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadfile
func (l *luaState) LoadFile(filename string) int {
	return l.LoadFileX(filename, "bt")
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadfilex
func (l *luaState) LoadFileX(filename, mode string) int {
	if data, err := ioutil.ReadFile(filename); err == nil {
		return l.Load(data, "@"+filename, mode)
	}
	return LUA_ERRFILE
}

// [-0, +1, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_loadstring
func (l *luaState) LoadString(s string) int {
	return l.Load([]byte(s), s, "bt")
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#luaL_typename
func (l *luaState) TypeName2(idx int) string {
	return l.TypeName(l.Type(idx))
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_len
func (l *luaState) Len2(idx int) int64 {
	l.Len(idx)
	i, isNum := l.ToIntegerX(-1)
	if !isNum {
		l.Error2("object length is not an integer")
	}
	l.Pop(1)
	return i
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_tolstring
func (l *luaState) ToString2(idx int) string {
	if l.CallMeta(idx, "__tostring") { /* metafield? */
		if !l.IsString(-1) {
			l.Error2("'__tostring' must return a string")
		}
	} else {
		switch l.Type(idx) {
		case LUA_TNUMBER:
			if l.IsInteger(idx) {
				l.PushString(fmt.Sprintf("%d", l.ToInteger(idx))) // todo
			} else {
				l.PushString(fmt.Sprintf("%g", l.ToNumber(idx))) // todo
			}
		case LUA_TSTRING:
			l.PushValue(idx)
		case LUA_TBOOLEAN:
			if l.ToBoolean(idx) {
				l.PushString("true")
			} else {
				l.PushString("false")
			}
		case LUA_TNIL:
			l.PushString("nil")
		default:
			tt := l.GetMetafield(idx, "__name") /* try name */
			var kind string
			if tt == LUA_TSTRING {
				kind = l.CheckString(-1)
			} else {
				kind = l.TypeName2(idx)
			}

			l.PushString(fmt.Sprintf("%s: %p", kind, l.ToPointer(idx)))
			if tt != LUA_TNIL {
				l.Remove(-2) /* remove '__name' */
			}
		}
	}
	return l.CheckString(-1)
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_getsubtable
func (l *luaState) GetSubTable(idx int, fname string) bool {
	if l.GetField(idx, fname) == LUA_TTABLE {
		return true /* table already there */
	}
	l.Pop(1) /* remove previous result */
	idx = l.stack.absIndex(idx)
	l.NewTable()
	l.PushValue(-1)        /* copy to be left at top */
	l.SetField(idx, fname) /* assign new table to field */
	return false           /* false, because did not find table there */
}

// [-0, +(0|1), m]
// http://www.lua.org/manual/5.3/manual.html#luaL_getmetafield
func (l *luaState) GetMetafield(obj int, event string) LuaType {
	if !l.GetMetatable(obj) { /* no metatable? */
		return LUA_TNIL
	}

	l.PushString(event)
	tt := l.RawGet(-2)
	if tt == LUA_TNIL { /* is metafield nil? */
		l.Pop(2) /* remove metatable and metafield */
	} else {
		l.Remove(-2) /* remove only metatable */
	}
	return tt /* return metafield type */
}

// [-0, +(0|1), e]
// http://www.lua.org/manual/5.3/manual.html#luaL_callmeta
func (l *luaState) CallMeta(obj int, event string) bool {
	obj = l.AbsIndex(obj)
	if l.GetMetafield(obj, event) == LUA_TNIL { /* no metafield? */
		return false
	}

	l.PushValue(obj)
	l.Call(1, 1)
	return true
}

// [-0, +0, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_openlibs
func (l *luaState) OpenLibs() {
	libs := map[string]GoFunction{
		"_G":        stdlib.OpenBaseLib,
		"math":      stdlib.OpenMathLib,
		"table":     stdlib.OpenTableLib,
		"string":    stdlib.OpenStringLib,
		"utf8":      stdlib.OpenUTF8Lib,
		"os":        stdlib.OpenOSLib,
		"package":   stdlib.OpenPackageLib,
		"coroutine": stdlib.OpenCoroutineLib,
	}

	for name, fun := range libs {
		l.RequireF(name, fun, true)
		l.Pop(1)
	}
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#luaL_requiref
func (l *luaState) RequireF(modname string, openf GoFunction, glb bool) {
	l.GetSubTable(LUA_REGISTRYINDEX, "_LOADED")
	l.GetField(-1, modname) /* LOADED[modname] */
	if !l.ToBoolean(-1) {   /* package not already loaded? */
		l.Pop(1) /* remove field */
		l.PushGoFunction(openf)
		l.PushString(modname)   /* argument to open function */
		l.Call(1, 1)            /* call 'openf' to open module */
		l.PushValue(-1)         /* make copy of module (call result) */
		l.SetField(-3, modname) /* _LOADED[modname] = module */
	}
	l.Remove(-2) /* remove _LOADED table */
	if glb {
		l.PushValue(-1)      /* copy of module */
		l.SetGlobal(modname) /* _G[modname] = module */
	}
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_newlib
func (l *luaState) NewLib(list FuncReg) {
	l.NewLibTable(list)
	l.SetFuncs(list, 0)
}

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_newlibtable
func (l *luaState) NewLibTable(list FuncReg) {
	l.CreateTable(0, len(list))
}

// [-nup, +0, m]
// http://www.lua.org/manual/5.3/manual.html#luaL_setfuncs
func (l *luaState) SetFuncs(list FuncReg, nup int) {
	l.CheckStack2(nup, "too many upvalues")
	for name, fun := range list { /* fill the table with given functions */
		for i := 0; i < nup; i++ { /* copy upvalues to the top */
			l.PushValue(-nup)
		}
		// r[-(nup+2)][name]=fun
		l.PushGoClosure(fun, nup) /* closure with those upvalues */
		l.SetField(-(nup + 2), name)
	}
	l.Pop(nup) /* remove upvalues */
}

func (l *luaState) intError(arg int) {
	if l.IsNumber(arg) {
		l.ArgError(arg, "number has no integer representation")
	} else {
		l.tagError(arg, LUA_TNUMBER)
	}
}

func (l *luaState) tagError(arg int, tag LuaType) {
	l.typeError(arg, l.TypeName(LuaType(tag)))
}

func (l *luaState) typeError(arg int, tname string) int {
	var typeArg string /* name for the type of the actual argument */
	if l.GetMetafield(arg, "__name") == LUA_TSTRING {
		typeArg = l.ToString(-1) /* use the given type name */
	} else if l.Type(arg) == LUA_TLIGHTUSERDATA {
		typeArg = "light userdata" /* special name for messages */
	} else {
		typeArg = l.TypeName2(arg) /* standard name */
	}
	msg := tname + " expected, got " + typeArg
	l.PushString(msg)
	return l.ArgError(arg, msg)
}
