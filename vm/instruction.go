package vm

import "luago/api"

type Instruction uint32

const MAXARG_Bx = 1<<18 - 1       // 2^18 - 1 = 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 262143 / 2 = 131071

/*
 31       22       13       5    0
  +-------+^------+-^-----+-^-----
  |b=9bits |c=9bits |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    bx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |   sbx=18bits    |a=8bits|op=6|
  +-------+^------+-^-----+-^-----
  |    ax=26bits            |op=6|
  +-------+^------+-^-----+-^-----
 31      23      15       7      0
*/

func (i Instruction) Opcode() int {
	return int(i & 0x3F) // 11 1111
}

func (i Instruction) ABC() (a, b, c int) {
	a = int(i >> 6 & 0xFF)   // 1111 1111
	c = int(i >> 14 & 0x1FF) // 1 1111 1111
	b = int(i >> 23 & 0x1FF)
	return
}

func (i Instruction) ABx() (a, bx int) {
	a = int(i >> 6 & 0xFF)
	bx = int(i >> 14)
	return
}

func (i Instruction) AsBx() (a, sbx int) {
	a, bx := i.ABx()
	return a, bx - MAXARG_sBx // sBx操作数（共18个比特）表示的是有符号整数。有很多种方式可以把有符号整数编码成比特序列，比如2的补码（Two’s Complement）等。Lua虚拟机这里采用了一种叫作偏移二进制码（Offset Binary，也叫作Excess-K）的编码模式。具体来说，如果把sBx解释成无符号整数时它的值是x，那么解释成有符号整数时它的值就是x-K。那么K是什么呢？K取sBx所能表示的最大无符号整数值的一半，也就是上面代码中的MAXARG_sBx。
}

func (i Instruction) Ax() int {
	return int(i >> 6)
}

func (i Instruction) OpName() string {
	return opcodes[i.Opcode()].name
}

func (i Instruction) OpMode() byte {
	return opcodes[i.Opcode()].opMode
}

func (i Instruction) BMode() byte {
	return opcodes[i.Opcode()].argBMode
}

func (i Instruction) CMode() byte {
	return opcodes[i.Opcode()].argCMode
}

func (i Instruction) Execute(vm api.LuaVM) {
	action := opcodes[i.Opcode()].action
	if action != nil {
		action(i, vm)
	} else {
		panic(i.OpName())
	}
}
