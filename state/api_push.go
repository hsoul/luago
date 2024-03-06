package state

import (
	"fmt"
	. "luago/api"
)

func (s *luaState) PushNil() {
	s.stack.push(nil)
}

func (s *luaState) PushBoolean(b bool) {
	s.stack.push(b)
}

func (s *luaState) PushInteger(n int64) {
	s.stack.push(n)
}

func (s *luaState) PushNumber(n float64) {
	s.stack.push(n)
}

func (s *luaState) PushString(str string) {
	s.stack.push(str)
}

// [-0, +1, e]
// http://www.lua.org/manual/5.3/manual.html#lua_pushfstring
func (self *luaState) PushFString(fmtStr string, a ...interface{}) {
	str := fmt.Sprintf(fmtStr, a...)
	self.stack.push(str)
}

func (s *luaState) PushGoFunction(f GoFunction) {
	s.stack.push(newGoClosure(f, 0))
}

func (s *luaState) PushGoClosure(f GoFunction, n int) {
	closure := newGoClosure(f, n)
	for i := 0; i < n; i++ {
		val := s.stack.pop()
		closure.upvals[i] = &upvalue{&val}
	}
	s.stack.push(closure)
}

func (s *luaState) PushGlobalTable() {
	global := s.registry.get(LUA_RIDX_GLOBALS)
	s.stack.push(global)
}

func (s *luaState) PushThread() bool {
	s.stack.push(s)
	return s.isMainThread()
}
