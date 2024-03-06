package vm

import . "luago/api"

// for 循环相关指令

// FORPREP
// R(A)-=R(A+2); pc+=sBx
func forPrep(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) -= R(A+2)
	vm.PushValue(a)
	vm.PushValue(a + 2)
	vm.Arith(LUA_OPSUB)
	vm.Replace(a)

	// pc+=sBx
	vm.AddPC(sBx)
}

// FORLOOP
// R(A) += R(A+2)
//
//	if R(A) <?= R(A+1) then {
//		  pc+=sBx; R(A+3)=R(A)
//	}
func forLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	// R(A) += R(A+2)
	vm.PushValue(a + 2)
	vm.PushValue(a)
	vm.Arith(LUA_OPADD)
	vm.Replace(a)

	// R(A) <?= R(A+1) // FORLOOP指令伪代码中的“<？=”符号。当步长是正数时，这个符号的含义是“<=”，也就是说继续循环的条件是数值不大于限制；当步长是负数时，这个符号的含义是“>=”，循环继续的条件就变成了数值不小于限制。
	isPositiveStep := vm.ToNumber(a+2) >= 0
	if isPositiveStep && vm.Compare(a, a+1, LUA_OPLE) ||
		!isPositiveStep && vm.Compare(a+1, a, LUA_OPLE) {

		// pc+=sBx; R(A+3)=R(A)
		vm.AddPC(sBx)
		vm.Copy(a, a+3)
	}
}

//	if R(A+1) ~= nil then {
//	    R(A)=R(A+1); pc += sBx
//	}
func tForLoop(i Instruction, vm LuaVM) {
	a, sBx := i.AsBx()
	a += 1

	if !vm.IsNil(a + 1) {
		vm.Copy(a+1, a)
		vm.AddPC(sBx)
	}
}
