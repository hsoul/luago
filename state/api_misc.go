package state

import "luago/number"

func (s *luaState) Len(idx int) {
	val := s.stack.get(idx)
	if str, ok := val.(string); ok {
		s.stack.push(int64(len(str)))
	} else if result, ok := callMetamethod(val, val, "__len", s); ok {
		s.stack.push(result)
	} else if t, ok := val.(*luaTable); ok {
		s.stack.push(int64(t.len()))
	} else {
		panic("length error!")
	}
}

func (s *luaState) Concat(n int) {
	if n == 0 {
		s.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if s.IsString(-1) && s.IsString(-2) {
				s2 := s.ToString(-1)
				s1 := s.ToString(-2)
				s.stack.pop()
				s.stack.pop()
				s.stack.push(s1 + s2)
				continue
			}

			b := s.stack.pop()
			a := s.stack.pop()
			if result, ok := callMetamethod(a, b, "__concat", s); ok {
				s.stack.push(result)
				continue
			}

			panic("concatenation error!")
		}
	}
	// n == 1, do nothing
}

// Next（）方法根据键获取表的下一个键值对。其中表的索引由参数指定，上一个键从栈顶弹出。
// 如果从栈顶弹出的键是nil，说明刚开始遍历表，把表的第一个键值对推入栈顶并返回true；
// 否则，如果遍历还没有结束，把下一个键值对推入栈顶并返回true；
// 如果表是空的，或者遍历已经结束，不用往栈里推入任何值，直接返回false即可。
func (s *luaState) Next(idx int) bool {
	val := s.stack.get(idx)
	if t, ok := val.(*luaTable); ok {
		key := s.stack.pop()
		if nextKey := t.nextKey(key); nextKey != nil {
			s.stack.push(nextKey)
			s.stack.push(t.get(nextKey))
			return true
		}
		return false
	}
	panic("table expected!")
}

func (s *luaState) Error() int {
	err := s.stack.pop()
	panic(err)
}

func (self *luaState) StringToNumber(s string) bool {
	if n, ok := number.ParseInteger(s); ok {
		self.PushInteger(n)
		return true
	}
	if n, ok := number.ParseFloat(s); ok {
		self.PushNumber(n)
		return true
	}
	return false
}
