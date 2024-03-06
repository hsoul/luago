package state

import . "luago/api"

type luaStack struct {
	// virtual stack
	slots []luaValue
	top   int // 记录栈顶索引，从1开始
	// call info
	state   *luaState
	closure *closure
	varargs []luaValue
	pc      int
	// linked list
	prev    *luaStack
	openuvs map[int]*upvalue
}

func newLuaStack(size int, state *luaState) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
		state: state,
	}
}

func (s *luaStack) check(n int) {
	free := len(s.slots) - s.top
	for i := free; i < n; i++ {
		s.slots = append(s.slots, nil)
	}
}

func (s *luaStack) push(val luaValue) {
	if s.top == len(s.slots) {
		panic("stack overflow!")
	}
	s.slots[s.top] = val
	s.top++
}

func (s *luaStack) pop() luaValue {
	if s.top < 1 {
		panic("stack underflow!")
	}
	s.top--
	val := s.slots[s.top]
	s.slots[s.top] = nil
	return val
}

func (s *luaStack) absIndex(idx int) int {
	if idx >= 0 || idx <= LUA_REGISTRYINDEX {
		return idx
	}
	return idx + s.top + 1
}

func (s *luaStack) isValid(idx int) bool {
	if idx < LUA_REGISTRYINDEX { // upvalues
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		closure := s.closure
		return closure != nil && uvIdx < len(closure.upvals)
	}
	if idx == LUA_REGISTRYINDEX {
		return true
	}
	absIdx := s.absIndex(idx)
	return absIdx > 0 && absIdx <= s.top
}

func (s *luaStack) get(idx int) luaValue {
	if idx < LUA_REGISTRYINDEX { // upvalues
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		closure := s.closure
		if closure == nil || uvIdx >= len(closure.upvals) {
			return nil
		}
		return *closure.upvals[uvIdx].val
	}
	if idx == LUA_REGISTRYINDEX {
		return s.state.registry
	}
	absIdx := s.absIndex(idx)
	if absIdx > 0 && absIdx <= s.top {
		return s.slots[absIdx-1]
	}
	return nil
}

func (s *luaStack) set(idx int, val luaValue) {
	if idx < LUA_REGISTRYINDEX { // upvalues
		uvIdx := LUA_REGISTRYINDEX - idx - 1
		closure := s.closure
		if closure != nil && uvIdx < len(closure.upvals) {
			*closure.upvals[uvIdx].val = val
		}
		return
	}
	if idx == LUA_REGISTRYINDEX {
		s.state.registry = val.(*luaTable)
		return
	}
	absIdx := s.absIndex(idx)
	if absIdx > 0 && absIdx <= s.top {
		s.slots[absIdx-1] = val
		return
	}
	panic("invalid index!")
}

func (s *luaStack) reverse(from, to int) {
	slots := s.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}

func (s *luaStack) popN(n int) []luaValue {
	vals := make([]luaValue, n)
	for i := n - 1; i >= 0; i-- {
		vals[i] = s.pop()
	}
	return vals
}

func (s *luaStack) pushN(vals []luaValue, n int) {
	nVals := len(vals)
	if n < 0 {
		n = nVals
	}
	for i := 0; i < n; i++ {
		if i < nVals {
			s.push(vals[i])
		} else {
			s.push(nil)
		}
	}
}
