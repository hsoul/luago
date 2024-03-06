package state

import . "luago/api"

func (s *luaState) GetTop() int {
	return s.stack.top
}

func (s *luaState) AbsIndex(idx int) int {
	return s.stack.absIndex(idx)
}

func (s *luaState) CheckStack(n int) bool {
	s.stack.check(n)
	return true // todo
}

func (s *luaState) Pop(n int) {
	for i := 0; i < n; i++ {
		s.stack.pop()
	}
}

// func (s *luaState) Pop(n int) {
// 	s.SetTop(-n - 1)
// }

func (s *luaState) Copy(fromIdx, toIdx int) {
	val := s.stack.get(fromIdx)
	s.stack.set(toIdx, val)
}

func (s *luaState) PushValue(idx int) {
	val := s.stack.get(idx)
	s.stack.push(val)
}

func (s *luaState) Replace(idx int) {
	val := s.stack.pop()
	s.stack.set(idx, val)
}

func (s *luaState) Insert(idx int) {
	s.Rotate(idx, 1)
}

func (s *luaState) Remove(idx int) {
	s.Rotate(idx, -1)
	s.Pop(1)
}

func (s *luaState) Rotate(idx, n int) {
	t := s.stack.top - 1
	p := s.stack.absIndex(idx) - 1
	var m int
	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	// fmt.Println("----- rotate start ------")
	// fmt.Printf("%d %d %d %d\n", t, p, m, n)
	s.stack.reverse(p, m)
	// fmt.Printf("----- rotate:reverse(%d, %d)------\n", p, m)
	// PrintStack(s)
	s.stack.reverse(m+1, t)
	// fmt.Printf("----- rotate:reverse(%d, %d)------\n", m+1, t)
	// PrintStack(s)
	s.stack.reverse(p, t)
	// fmt.Printf("----- rotate:reverse(%d, %d)------\n", p, t)
	// PrintStack(s)
	// fmt.Println("----- rotate end ------")
}

func (s *luaState) SetTop(idx int) {
	newTop := s.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}
	n := s.stack.top - newTop
	if n > 0 {
		s.Pop(n)
	} else if n < 0 {
		for i := 0; i > n; i-- {
			s.stack.push(nil)
		}
	}
}

func (s *luaState) XMove(to LuaState, n int) {
	vals := s.stack.popN(n)
	to.(*luaState).stack.pushN(vals, n)
}
