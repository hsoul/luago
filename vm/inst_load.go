package vm

import . "luago/api"

// 加载类指令

// R(A), R(A+1), ..., R(A+B) := nil
func loadNil(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	vm.PushNil()
	for i := a; i <= a+b; i++ {
		vm.Copy(-1, i)
	}
	vm.Pop(1)
}

// LOADBOOL指令（iABC模式）给单个寄存器设置布尔值。寄存器索引由操作数A指定，布尔值由寄存器B指定（0代表false，非0代表true），如果寄存器C非0则跳过下一条指令
// R(A) := (bool)B; if (C) pc++
// luac -l -
// local a,b,c,d,e; c = a > b
// 解释下指令
func loadBool(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	vm.PushBoolean(b != 0)
	vm.Replace(a)
	if c != 0 {
		vm.AddPC(1)
	}
}

// LOADK指令（iABx模式）将常量表里的某个常量加载到指定寄存器，寄存器索引由操作数A指定，常量表索引由操作数Bx指定。
// 如果用Kst（N）表示常量表中的第N个常量，那么LOADK指令可以用以下伪代码表示
// R(A) := Kst(Bx)
func loadK(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1

	vm.GetConst(bx)
	vm.Replace(a)
}

// 我们知道操作数Bx占18个比特，能表示的最大无符号整数是262143，大部分Lua函数的常量表大小都不会超过这个数，所以这个限制通常不是什么问题。不过Lua也经常被当作数据描述语言使用，所以常量表大小可能超过这个限制也并不稀奇。为了应对这种情况，Lua还提供了一条LOADKX指令。
// LOADKX指令（也是iABx模式）需要和EXTRAARG指令（iAx模式）搭配使用，用后者的Ax操作数来指定常量索引。Ax操作数占26个比特，可以表达的最大无符号整数是67108864
func loadKx(i Instruction, vm LuaVM) {
	a, _ := i.ABx()
	a += 1

	ax := Instruction(vm.Fetch()).Ax()

	vm.GetConst(ax)
	vm.Replace(a)
}
