package vm

import "luago/api"

// 每条Lua虚拟机指令占用4个字节，共32个比特（可以用Go语言uint32类型表示），其中低6个比特用于操作码，高26个比特用于操作数。
// 按照高26个比特的分配（以及解释）方式，Lua虚拟机指令可以分为四类，分别对应四种编码模式（Mode）：iABC、iABx、iAsBx、iAx。

// 在4种模式中，只有iAsBx模式下的sBx操作数会被解释成有符号整数，其他情况下操作数均被解释为无符号整数。

// Lua虚拟机指令以iABC模式居多，在总计47条指令中，有39条使用iABC模式。其余8条指令中，有3条使用iABx指令，4条使用iAsBx模式，1条使用iAx格式

const (
	IABC  = iota // iABC模式的指令可以携带A、B、C三个操作数，分别占用8、9、9个比特；
	IABx         // iABx模式的指令可以携带A和Bx两个操作数，分别占用8和18个比特；
	IAsBx        // iAsBx模式的指令可以携带A和sBx两个操作数，分别占用8和18个比特；
	IAx          // iAx模式的指令只携带一个操作数，占用全部的26个比特。
)

const (
	OP_MOVE = iota
	OP_LOADK
	OP_LOADKX
	OP_LOADBOOL
	OP_LOADNIL
	OP_GETUPVAL
	OP_GETTABUP
	OP_GETTABLE
	OP_SETTABUP
	OP_SETUPVAL
	OP_SETTABLE
	OP_NEWTABLE
	OP_SELF
	OP_ADD
	OP_SUB
	OP_MUL
	OP_MOD
	OP_POW
	OP_DIV
	OP_IDIV
	OP_BAND
	OP_BOR
	OP_BXOR
	OP_SHL
	OP_SHR
	OP_UNM
	OP_BNOT
	OP_NOT
	OP_LEN
	OP_CONCAT
	OP_JMP
	OP_EQ
	OP_LT
	OP_LE
	OP_TEST
	OP_TESTSET
	OP_CALL
	OP_TAILCALL
	OP_RETURN
	OP_FORLOOP
	OP_FORPREP
	OP_TFORCALL
	OP_TFORLOOP
	OP_SETLIST
	OP_CLOSURE
	OP_VARARG
	OP_EXTRAARG
)

const (
	OpArgN = iota // 操作数不使用 // 在iABC模式下，B和C操作数各占9个比特，如果B或C操作数属于OpArgK类型，那么就只能使用9个比特中的低8位，最高位的那个比特如果是1，则操作数表示常量表索引，否则表示寄存器索引。
	OpArgU        //
	OpArgR        // 操作数是寄存器或跳转偏移量 // 在iABC模式下，B和C操作数各占9个比特，如果B或C操作数属于OpArgK类型，那么就只能使用9个比特中的低8位，最高位的那个比特如果是1，则操作数表示常量表索引，否则表示寄存器索引。
	OpArgK        // 操作数是常量表索引或常量索引/寄存器索引 // 在iABC模式下，B和C操作数各占9个比特，如果B或C操作数属于OpArgK类型，那么就只能使用9个比特中的低8位，最高位的那个比特如果是1，则操作数表示常量表索引，否则表示寄存器索引。
)

type opcode struct {
	testFlag byte // 指令是否是测试指令(下一条指令一定是跳转指令)
	setAFlag byte // 指令是否修改了寄存器A
	argBMode byte // 指令B操作数的模式
	argCMode byte // 指令C操作数的模式
	opMode   byte // 指令模式
	name     string
	action   func(i Instruction, vm api.LuaVM)
}

var opcodes = []opcode{
	/*     T  A    B       C     mode    name    action*/
	opcode{0, 1, OpArgR, OpArgN, IABC, "MOVE    ", move},
	opcode{0, 1, OpArgK, OpArgN, IABx, "LOADK   ", loadK},
	opcode{0, 1, OpArgN, OpArgN, IABx, "LOADKX  ", loadKx},
	opcode{0, 1, OpArgU, OpArgU, IABC, "LOADBOOL", loadBool},
	opcode{0, 1, OpArgU, OpArgN, IABC, "LOADNIL ", loadNil},
	opcode{0, 1, OpArgU, OpArgN, IABC, "GETUPVAL", getUpVal},
	opcode{0, 1, OpArgU, OpArgK, IABC, "GETTABUP", getTabUp},
	opcode{0, 1, OpArgR, OpArgK, IABC, "GETTABLE", getTable},
	opcode{0, 0, OpArgK, OpArgK, IABC, "SETTABUP", setTabUp},
	opcode{0, 0, OpArgU, OpArgN, IABC, "SETUPVAL", setUpVal},
	opcode{0, 0, OpArgK, OpArgK, IABC, "SETTABLE", setTable},
	opcode{0, 1, OpArgU, OpArgU, IABC, "NEWTABLE", newTable},
	opcode{0, 1, OpArgR, OpArgK, IABC, "SELF    ", self},
	opcode{0, 1, OpArgK, OpArgK, IABC, "ADD     ", add},
	opcode{0, 1, OpArgK, OpArgK, IABC, "SUB     ", sub},
	opcode{0, 1, OpArgK, OpArgK, IABC, "MUL     ", mul},
	opcode{0, 1, OpArgK, OpArgK, IABC, "MOD     ", mod},
	opcode{0, 1, OpArgK, OpArgK, IABC, "POW     ", pow},
	opcode{0, 1, OpArgK, OpArgK, IABC, "DIV     ", div},
	opcode{0, 1, OpArgK, OpArgK, IABC, "IDIV    ", idiv},
	opcode{0, 1, OpArgK, OpArgK, IABC, "BAND    ", band},
	opcode{0, 1, OpArgK, OpArgK, IABC, "BOR     ", bor},
	opcode{0, 1, OpArgK, OpArgK, IABC, "BXOR    ", bxor},
	opcode{0, 1, OpArgK, OpArgK, IABC, "SHL     ", shl},
	opcode{0, 1, OpArgK, OpArgK, IABC, "SHR     ", shr},
	opcode{0, 1, OpArgR, OpArgN, IABC, "UNM     ", unm},
	opcode{0, 1, OpArgR, OpArgN, IABC, "BNOT    ", bnot},
	opcode{0, 1, OpArgR, OpArgN, IABC, "NOT     ", not},
	opcode{0, 1, OpArgR, OpArgN, IABC, "LEN     ", length},
	opcode{0, 1, OpArgR, OpArgR, IABC, "CONCAT  ", concat},
	opcode{0, 0, OpArgR, OpArgN, IAsBx, "JMP     ", jmp},
	opcode{1, 0, OpArgK, OpArgK, IABC, "EQ      ", eq},
	opcode{1, 0, OpArgK, OpArgK, IABC, "LT      ", lt},
	opcode{1, 0, OpArgK, OpArgK, IABC, "LE      ", le},
	opcode{1, 0, OpArgN, OpArgU, IABC, "TEST    ", test},
	opcode{1, 1, OpArgR, OpArgU, IABC, "TESTSET ", testSet},
	opcode{0, 1, OpArgU, OpArgU, IABC, "CALL    ", call},
	opcode{0, 1, OpArgU, OpArgU, IABC, "TAILCALL", tailCall},
	opcode{0, 0, OpArgU, OpArgN, IABC, "RETURN  ", _reutrn},
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "FORLOOP ", forLoop},
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "FORPREP ", forPrep},
	opcode{0, 0, OpArgN, OpArgU, IABC, "TFORCALL", tForCall},
	opcode{0, 1, OpArgR, OpArgN, IAsBx, "TFORLOOP", tForLoop},
	opcode{0, 0, OpArgU, OpArgU, IABC, "SETLIST ", setList},
	opcode{0, 1, OpArgU, OpArgN, IABx, "CLOSURE ", closure},
	opcode{0, 1, OpArgU, OpArgN, IABC, "VARARG  ", vararg},
	opcode{0, 0, OpArgU, OpArgU, IAx, "EXTRAARG", nil},
}
