package state

import (
	"fmt"
	. "luago/api"
)

func (s *luaState) TypeName(tp LuaType) string {
	switch tp {
	case LUA_TNONE:
		return "no value"
	case LUA_TNIL:
		return "nil"
	case LUA_TBOOLEAN:
		return "boolean"
	case LUA_TNUMBER:
		return "number"
	case LUA_TSTRING:
		return "string"
	case LUA_TTABLE:
		return "table"
	case LUA_TFUNCTION:
		return "function"
	case LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}

func (s *luaState) Type(idx int) LuaType {
	if s.stack.isValid(idx) {
		val := s.stack.get(idx)
		return typeOf(val)
	}
	return LUA_TNONE
}

func (s *luaState) IsNone(idx int) bool {
	return s.Type(idx) == LUA_TNONE
}

func (s *luaState) IsNil(idx int) bool {
	return s.Type(idx) == LUA_TNIL
}

func (s *luaState) IsNoneOrNil(idx int) bool {
	return s.Type(idx) <= LUA_TNIL
}

func (s *luaState) IsBoolean(idx int) bool {
	return s.Type(idx) == LUA_TBOOLEAN
}

func (s *luaState) IsString(idx int) bool {
	t := s.Type(idx)
	return t == LUA_TSTRING || t == LUA_TNUMBER
}

func (s *luaState) IsTable(idx int) bool {
	return s.Type(idx) == LUA_TTABLE
}

func (s *luaState) IsThread(idx int) bool {
	return s.Type(idx) == LUA_TTHREAD
}

func (s *luaState) IsFunction(idx int) bool {
	return s.Type(idx) == LUA_TFUNCTION
}

func (s *luaState) IsNumber(idx int) bool {
	_, ok := s.ToNumberX(idx)
	return ok
}

func (s *luaState) IsInteger(idx int) bool {
	val := s.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

func (s *luaState) ToBoolean(idx int) bool {
	val := s.stack.get(idx)
	return convertToBoolean(val)
}

func (s *luaState) ToNumber(idx int) float64 {
	n, _ := s.ToNumberX(idx)
	return n
}

func (s *luaState) ToNumberX(idx int) (float64, bool) {
	val := s.stack.get(idx)
	return convertToFloat(val)
}

func (s *luaState) ToInteger(idx int) int64 {
	i, _ := s.ToIntegerX(idx)
	return i
}

func (s *luaState) ToIntegerX(idx int) (int64, bool) {
	val := s.stack.get(idx)
	return convertToInteger(val)
}

func (s *luaState) ToStringX(idx int) (string, bool) {
	val := s.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		str := fmt.Sprintf("%v", x)
		s.stack.set(idx, str) // 这里会修改栈中的值
		return str, true
	default:
		return "", false
	}
}

func (s *luaState) ToString(idx int) string {
	str, _ := s.ToStringX(idx)
	return str
}

func (s *luaState) IsGoFunction(idx int) bool {
	val := s.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc != nil
	}
	return false
}

func (s *luaState) ToGoFunction(idx int) GoFunction {
	val := s.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}
	return nil
}

func (s *luaState) RawLen(idx int) uint {
	val := s.stack.get(idx)
	switch x := val.(type) {
	case string:
		return uint(len(x))
	case *luaTable:
		return uint(x.len())
	default:
		return 0
	}
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_topointer
func (self *luaState) ToPointer(idx int) interface{} {
	// todo
	return self.stack.get(idx)
}

func (s *luaState) ToThread(idx int) LuaState {
	val := s.stack.get(idx)
	if val != nil {
		if ls, ok := val.(*luaState); ok {
			return ls
		}
	}
	return nil
}

func (s *luaState) PrintValue(idx int) {
	t := s.Type(idx)
	switch t {
	case LUA_TBOOLEAN:
		fmt.Printf("[%t]", s.ToBoolean(idx))
	case LUA_TNUMBER:
		fmt.Printf("[%g]", s.ToNumber(idx))
	case LUA_TSTRING:
		fmt.Printf("[%q]", s.ToString(idx))
	case LUA_TTABLE:
		PrintTable(s.stack.get(idx).(*luaTable))
	default:
		fmt.Printf("[%s]", s.TypeName(t))
	}
	fmt.Println()
}
