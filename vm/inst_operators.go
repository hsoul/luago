package vm

import . "luago/api"

// 运算符相关指令

// R(A) := RK(B) + RK(C)
func _binaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, c := i.ABC()
	a += 1

	vm.GetRK(b)
	vm.GetRK(c)
	vm.Arith(op)
	vm.Replace(a)
}

// R(A) := op R(B)
func _unaryArith(i Instruction, vm LuaVM, op ArithOp) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.PushValue(b)
	vm.Arith(op)
	vm.Replace(a)
}

func add(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPADD) }  // +
func sub(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSUB) }  // -
func mul(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMUL) }  // *
func mod(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPMOD) }  // %
func pow(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPPOW) }  // ^
func div(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPDIV) }  // /
func idiv(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPIDIV) } // //
func band(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBAND) } // &
func bor(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPBOR) }  // |
func bxor(i Instruction, vm LuaVM) { _binaryArith(i, vm, LUA_OPBXOR) } // ~
func shl(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHL) }  // <<
func shr(i Instruction, vm LuaVM)  { _binaryArith(i, vm, LUA_OPSHR) }  // >>
func unm(i Instruction, vm LuaVM)  { _unaryArith(i, vm, LUA_OPUNM) }   // - (unary minus)
func bnot(i Instruction, vm LuaVM) { _unaryArith(i, vm, LUA_OPBNOT) }  // ~
func eq(i Instruction, vm LuaVM)   { _compare(i, vm, LUA_OPEQ) }       // ==
func lt(i Instruction, vm LuaVM)   { _compare(i, vm, LUA_OPLT) }       // <
func le(i Instruction, vm LuaVM)   { _compare(i, vm, LUA_OPLE) }       // <=

// R(A) := length of R(B)
func length(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.Len(b)
	vm.Replace(a)
}

// R(A) := R(B).. ... ..R(C)
func concat(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1
	c += 1

	n := c - b + 1
	vm.CheckStack(n) // 检查栈空间是否足够，如果不够，就扩容
	for i := b; i <= c; i++ {
		vm.PushValue(i)
	}
	vm.Concat(n) // 栈顶存放的是连接后的字段，栈多了一个元素
	vm.Replace(a)
}

// 比较指令（iABC模式），比较寄存器或常量表里的两个值（索引分别由操作数B和C指定），如果比较结果和操作数A（转换为布尔值）匹配，则跳过下一条指令。比较指令不改变寄存器状态
// 当A为0时，RK(B)和RK(C)的值相等，PC自增1。当A为1时，RK(B)和RK(C)不相等，PC自增1
// if ((RK(B) op RK(C)) ~= A) then pc++
// local a,b,c,d,e; a = (b == "foo")
func _compare(i Instruction, vm LuaVM, op CompareOp) {
	a, b, c := i.ABC()
	vm.GetRK(b)
	vm.GetRK(c)
	if vm.Compare(-2, -1, op) != (a != 0) {
		vm.AddPC(1)
	}
	vm.Pop(2)
}

// R(A) := not R(B)
func not(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1

	vm.PushBoolean(!vm.ToBoolean(b))
	vm.Replace(a)
}

// TESTSET指令（iABC模式），判断寄存器B（索引由操作数B指定）中的值转换为布尔值之后是否和操作数C表示的布尔值一致，如果一致则将寄存器B中的值复制到寄存器A（索引由操作数A指定）中，否则跳过下一条指令
// if (R(B) <=> C) then R(A) := R(B) else pc++
// local a,b,c,d,e; b = d and e
func testSet(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	if vm.ToBoolean(b) == (c != 0) {
		vm.Copy(b, a)
	} else {
		vm.AddPC(1)
	}
}

// TEST指令（iABC模式），判断寄存器A（索引由操作数A指定）中的值转换为布尔值之后是否和操作数C表示的布尔值一致，如果一致，则跳过下一条指令。TEST指令不使用操作数B，也不改变寄存器状态，可以用以下伪代码表示。
// if not (R(A) <=> C) then pc++
// local a,b,c,d,e; b = b and e
func test(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1

	if vm.ToBoolean(a) != (c != 0) {
		vm.AddPC(1)
	}
}
