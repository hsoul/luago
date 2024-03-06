package vm

import . "luago/api"

// 混杂指令

// R(A) := R(B)
func move(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1
	b += 1
	vm.Copy(b, a)
}

// JMP指令除了可以进行无条件跳转之外，还兼顾着闭合处于开启状态的Upvalue的责任。
// 如果某个块内部定义的局部变量已经被嵌套函数捕获，那么当这些局部变量退出作用域（也就是块结束）时，编译器会生成一条JMP指令，指示虚拟机闭合相应的Upvalue。
// pc+=sBx; if (A) close all upvalues >= R(A - 1)
func jmp(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	vm.AddPC(sBx)
	if a != 0 {
		vm.CloseUpvalues(a)
	}
}
