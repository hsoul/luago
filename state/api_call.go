package state

import (
	. "luago/api"
	"luago/binchunk"
	"luago/compiler"
	"luago/vm"
)

func (s *luaState) Load(chunk []byte, chunkName, mode string) int {
	var proto *binchunk.Prototype
	if binchunk.IsBinaryChunk(chunk) {
		proto = binchunk.Undump(chunk)
	} else {
		proto = compiler.Compile(string(chunk), chunkName)
		binchunk.List(proto)
	}

	// return 0

	c := newLuaClosure(proto)
	s.stack.push(c)
	if len(proto.Upvalues) > 0 { // 设置 _ENV
		env := s.registry.get(LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return 0
}

func (s *luaState) Call(nArgs, nResults int) {
	val := s.stack.get(-(nArgs + 1)) // 获取被调函数

	c, ok := val.(*closure)
	if !ok {
		if mf := getMetafield(val, "__call", s); mf != nil {
			if c, ok = mf.(*closure); ok {
				s.stack.push(val)
				s.Insert(-(nArgs + 2))
				nArgs += 1
			}
		}
	}
	if ok {
		if c.proto != nil { // 调用lua函数
			// fmt.Printf("call %s<%d,%d>\n", c.proto.Source, c.proto.LineDefined, c.proto.LastLineDefined)
			s.callLuaClosure(nArgs, nResults, c)
		} else { // 调用go函数
			// funcPtr := runtime.FuncForPC(reflect.ValueOf(c.goFunc).Pointer())
			// fmt.Printf("call %s<%v>\n", funcPtr.Name(), c.goFunc)
			// PrintStack(s)
			s.callGoClosure(nArgs, nResults, c)
		}
		// PrintStack(s)
	} else {
		panic("not function!")
	}
}

func (s *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize) // 函数执行需要寄存器数量
	nParams := int(c.proto.NumParams)  // 函数固定参数数量
	isVararg := c.proto.IsVararg == 1  // 是否是vararg函数

	newStack := newLuaStack(nRegs+LUA_MINSTACK, s) // 创建了一个新的调用帧，它的寄存器数量是被调函数需要的寄存器数量加上20个额外的空闲寄存器
	newStack.closure = c

	funcAndArgs := s.stack.popN(nArgs + 1) // 新的调用帧创建好之后，我们调用当前帧的popN（）方法把函数和参数值一次性从栈顶弹出，然后调用新帧的pushN（）方法按照固定参数数量传入参数
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs             // 固定参数传递完毕之后，需要修改新帧的栈顶指针，让它指向最后一个寄存器
	if nArgs > nParams && isVararg { // 如果被调函数是vararg函数，且传入参数的数量多于固定参数数量，还需要把vararg参数记下来，存在调用帧里，以备后用
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	s.pushLuaStack(newStack) // 我们把新调用帧推入调用栈顶，让它成为当前帧，然后调用runLuaClosure（）方法执行被调函数的指令。
	s.runLuaClosure()        // 指令执行完毕之后，新调用帧的使命就结束了，把它从调用栈顶弹出，这样主调帧就又成了当前帧。被调函数运行完毕之后，返回值会留在被调帧的栈顶（寄存器之上）
	s.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		s.stack.check(len(results))
		s.stack.pushN(results, nResults)
	}
}

func (s *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(s.Fetch())
		// fmt.Printf("[%02d] %s\n", s.stack.pc, inst.OpName())
		// PrintStack(s)
		inst.Execute(s)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}

func (s *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+LUA_MINSTACK, s)
	newStack.closure = c

	args := s.stack.popN(nArgs)
	newStack.pushN(args, nArgs)
	s.stack.pop() // pop closure

	s.pushLuaStack(newStack)
	r := c.goFunc(s) // 调用go函数
	s.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(r)
		s.stack.check(len(results))
		s.stack.pushN(results, nResults)
	}
}

func (s *luaState) PCall(nArgs, nResults, msgh int) (status int) {
	caller := s.stack
	status = LUA_ERRRUN

	defer func() {
		if err := recover(); err != nil {
			for s.stack != caller {
				s.popLuaStack()
			}
			s.stack.push(err)
		}
	}()

	s.Call(nArgs, nResults)
	status = LUA_OK
	return
}
