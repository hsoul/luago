package vm

import (
	. "luago/api"
)

// R(A) := closure(KPROTO[Bx])
func closure(i Instruction, vm LuaVM) {
	a, bx := i.ABx()
	a += 1

	vm.LoadProto(bx)
	vm.Replace(a)
}

// CALL指令可以借助前面介绍的Call（）方法实现。我们先调用_pushFuncAndArgs（）函数把被调函数和参数值推入栈顶，然后让Call（）方法去处理函数调用逻辑。
// Call（）方法结束之后，函数返回值已经在栈顶，调用_popResults（）函数把这些返回值移动到适当的寄存器中就可以了。
// R(A), ... ,R(A+C-2) := R(A)(R(A+1), ... ,R(A+B-1))

// 被调用函数位于寄存器中（索引由 A 指定），传递给被调用函数的参数值也在寄存器中，紧挨着被调用函数，参数个数为操作数 B 指定。
// ① B==0，接受其他函数全部返回来的参数
// ② B>0，参数个数为 B-1

// 函数调用结束后，原先存放函数和参数值的寄存器会被返回值占据，具体多少个返回值由操作数 C 指定。
// ① C==0，将返回值全部返回给接收者
// ② C==1，无返回值
// ③ C>1，返回值的数量为 C-1

func call(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	// fmt.Printf("a:%d, b:%d, c:%d\n", a, b, c)
	a += 1

	nArgs := _pushFuncAndArgs(a, b, vm) // 将函数和参数依次推入栈顶
	vm.Call(nArgs, c-1)                 // 返回值的数量是c-1
	_popResults(a, c, vm)               // 栈 -> 寄存器
}

func _pushFuncAndArgs(a, b int, vm LuaVM) (nArgs int) {
	if b >= 1 { // b-1个参数
		vm.CheckStack(b)
		for i := a; i < a+b; i++ { // 将寄存器中的参数推入栈顶
			vm.PushValue(i)
		}
		return b - 1
	} else {
		_fixStack(a, vm)
		return vm.GetTop() - vm.RegisterCount() - 1
	}
}

// 如果操作数C大于1，则返回值数量是C-1，循环调用Replace（）方法把栈顶返回值移动到相应寄存器即可；
// 如果操作数C等于1，则返回值数量是0，不需要任何处理；
// 如果C等于0，那么需要把被调函数的返回值全部返回。对于最后这种情况，干脆就把这些返回值先留在栈顶，反正后面也是要把它们再推入栈顶的。
// 我们往栈顶推入一个整数值，标记这些返回值原本是要移动到哪些寄存器中。
func _popResults(a, c int, vm LuaVM) {
	if c == 1 { // no results
	} else if c > 1 { // c-1个结果
		for i := a + c - 2; i >= a; i-- {
			vm.Replace(i)
		}
	} else { // c == 0, pop all results
		vm.CheckStack(1)
		vm.PushInteger(int64(a))
	}
}

func _fixStack(a int, vm LuaVM) {
	x := int(vm.ToInteger(-1))
	vm.Pop(1)

	vm.CheckStack(x - a)
	for i := a; i < x; i++ {
		vm.PushValue(i)
	}
	vm.Rotate(vm.RegisterCount()+1, x-a)
}

// 我们需要将返回值推入栈顶。如果操作数B等于1，则不需要返回任何值；
// 如果操作数B大于1，则需要返回B-1个值，这些值已经在寄存器里了，循环调用PushValue（）方法复制到栈顶即可。
// 如果操作数B等于0，则一部分返回值已经在栈顶了，调用_fixStack（）函数把另一部分也推入栈顶
// return R(A), ... ,R(A+B-2)
func _reutrn(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b == 1 {
		// no return values
	} else if b > 1 {
		// b-1 return values
		vm.CheckStack(b - 1)
		for i := a; i <= a+b-2; i++ {
			vm.PushValue(i)
		}
	} else {
		_fixStack(a, vm)
	}
}

// 操作数B若大于1，表示把B-1个vararg参数复制到寄存器；
// 否则只能等于0，表示把全部vararg参数复制到寄存器。
// 对于这两种情况，我们统一调用LoadVararg（）方法把vararg参数推入栈顶，剩下的工作交给_popResults（）函数就可以了。
// R(A), R(A+1), ..., R(A+B-2) = vararg
func vararg(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	if b != 1 { // b == 0 or b > 1
		vm.LoadVararg(b - 1)  // 加载变长参数到函数栈顶
		_popResults(a, b, vm) // 有什么用？
	}
}

// 函数调用一般通过调用栈来实现。用这种方法，每调用一个函数都会产生一个调用帧。
// 如果方法调用层次太深（特别是递归调用函数时），就容易导致调用栈溢出。
// 那么，有没有一种技术，既能让我们发挥递归函数的威力，又能避免调用栈溢出呢？有，那就是尾递归优化。
// 利用这种优化，被调函数可以重用主调函数的调用帧，因此可以有效缓解调用栈溢出症状。
// 不过尾递归优化只适用于某些特定的情况，并不能包治百病。
// 我们只要知道return f（args）这样的返回语句会被Lua编译器编译成TAILCALL指令就可以了
// return R(A)(R(A+1), ... ,R(A+B-1))
func tailCall(i Instruction, vm LuaVM) {
	a, b, _ := i.ABC()
	a += 1

	c := 0
	nArgs := _pushFuncAndArgs(a, b, vm)
	vm.Call(nArgs, c-1)
	_popResults(a, c, vm)
}

// SELF指令（iABC模式）把对象和方法拷贝到相邻的两个目标寄存器中。对象在寄存器中，索引由操作数B指定。方法名在常量表里，索引由操作数C指定。目标寄存器索引由操作数A指定。
// R(A+1) := R(B); R(A) := R(B)[RK(C)]
func self(i Instruction, vm LuaVM) {
	a, b, c := i.ABC()
	a += 1
	b += 1

	vm.Copy(b, a+1)
	vm.GetRK(c)
	vm.GetTable(b)
	vm.Replace(a)
}

// R(A+3), ... ,R(A+2+C) := R(A)(R(A+1), R(A+2));
func tForCall(i Instruction, vm LuaVM) {
	a, _, c := i.ABC()
	a += 1

	_pushFuncAndArgs(a, 3, vm)
	vm.Call(2, c)
	_popResults(a+3, c+1, vm)
}
