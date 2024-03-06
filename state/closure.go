package state

import (
	. "luago/api"
	"luago/binchunk"
)

type upvalue struct {
	val *luaValue
}

// 对于每个Upvalue，又有两种情况需要考虑：如果某一个Upvalue捕获的是当前函数的局部变量（Instack==1），那么我们只要访问当前函数的局部变量即可；
// 如果某一个Upvalue捕获的是更外围的函数中的局部变量（Instack==0），该Upvalue已经被当前函数捕获，我们只要把该Upvalue传递给闭包即可。

// 对于第一种情况，如果Upvalue捕获的外围函数局部变量还在栈上，直接引用即可，我们称这种Upvalue处于开放（Open）状态；
// 反之，必须把变量的实际值保存在其他地方，我们称这种Upvalue处于闭合（Closed）状态。

// 为了能够在合适的时机（比如局部变量退出作用域时，详见10.3.5节）把处于开放状态的Upvalue闭合，需要记录所有暂时还处于开放状态的Upvalue，我们把这些Upvalue记录在被捕获局部变量所在的栈帧里。
// 请读者打开luaStack.go文件（和closure.go文件在同一目录下），给luaStack结构体添加openuvs字段。该字段是map类型，其中键是int类型，存放局部变量的寄存器索引，值是Upvalue指针。

type closure struct {
	proto  *binchunk.Prototype // lua closure
	goFunc GoFunction          // go closure
	upvals []*upvalue
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	c := &closure{proto: proto}
	if nUpvals := len(proto.Upvalues); nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}

func newGoClosure(f GoFunction, nUpvals int) *closure {
	c := &closure{goFunc: f}
	if nUpvals > 0 {
		c.upvals = make([]*upvalue, nUpvals)
	}
	return c
}
