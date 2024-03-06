package state

func (s *luaState) PC() int {
	return s.stack.pc
}

func (s *luaState) AddPC(n int) {
	s.stack.pc += n
}

func (s *luaState) Fetch() uint32 {
	i := s.stack.closure.proto.Code[s.stack.pc]
	s.stack.pc++
	return i
}

func (s *luaState) GetConst(idx int) {
	c := s.stack.closure.proto.Constants[idx]
	s.stack.push(c)
}

func (s *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		s.GetConst(rk & 0xFF)
	} else { // register
		s.PushValue(rk + 1) // 传递给GetRK（）方法的参数实际上是iABC模式指令里的OpArgK类型参数。由第3章可知，这种类型的参数一共占9个比特。如果最高位是1，那么参数里存放的是常量表索引，把最高位去掉就可以得到索引值；否则最高位是0，参数里存放的就是寄存器索引值。但是请读者留意，Lua虚拟机指令操作数里携带的寄存器索引是从0开始的，而Lua API里的栈索引是从1开始的，所以当需要把寄存器索引当成栈索引使用时，要对寄存器索引加1。
	}
}

func (s *luaState) RegisterCount() int {
	return int(s.stack.closure.proto.MaxStackSize)
}

func (s *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(s.stack.varargs)
	}

	s.stack.check(n)
	s.stack.pushN(s.stack.varargs, n)
}

func (s *luaState) LoadProto(idx int) {
	subProto := s.stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	s.stack.push(closure)

	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instack == 1 {
			if s.stack.openuvs == nil {
				s.stack.openuvs = make(map[int]*upvalue)
			}
			if openuv, found := s.stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				closure.upvals[i] = &upvalue{&s.stack.slots[uvIdx]}
				s.stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {
			closure.upvals[i] = s.stack.closure.upvals[uvIdx]
		}
	}
}

func (s *luaState) CloseUpvalues(a int) {
	for i, openuv := range s.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(s.stack.openuvs, i) // 处于开启状态的Upvalue引用了还在寄存器里的Lua值，我们把这些Lua值从寄存器里复制出来，然后更新Upvalue，这样就将其改为了闭合状态。
		}
	}
}
