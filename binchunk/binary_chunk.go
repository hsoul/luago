package binchunk

const (
	LUA_SIGNATURE    = "\x1bLua"
	LUAC_VERSION     = 0x53
	LUAC_FORMAT      = 0
	LUAC_DATA        = "\x19\x93\r\n\x1a\n"
	CINT_SIZE        = 4
	CSIZET_SIZE      = 8
	INSTRUCTION_SIZE = 4
	LUA_INTEGER_SIZE = 8
	LUA_NUMBER_SIZE  = 8
	LUAC_INT         = 0x5678
	LUAC_NUM         = 370.5
)

const (
	TAG_NIL       = 0x00
	TAG_BOOLEAN   = 0x01
	TAG_NUMBER    = 0x03
	TAG_INTEGER   = 0x13
	TAG_SHORT_STR = 0x04
	TAG_LONG_STR  = 0x14
)

type binaryChunk struct {
	header
	sizeUpvalues byte       // 主函数upvalue数量
	mainFunc     *Prototype // 主函数原型
}

type header struct {
	signature       [4]byte // 0x1B4C7561 魔数主要起快速识别文件格式的作用
	version         byte    // lua 版本号，计算公式：major * 16 + minor 不计算发布号是因为发布号只是修复bug，不会改变文件格式
	format          byte    // 版本号之后的一个字节记录二进制chunk格式号。Lua虚拟机在加载二进制chunk时，会检查其格式号，如果和虚拟机本身的格式号不匹配，就拒绝加载该文件。Lua官方实现使用的格式号是0，如下所示
	luacData        [6]byte // 格式号之后的6个字节在Lua官方实现里叫作LUAC_DATA。其中前两个字节是0x1993，这是Lua 1.0发布的年份；后四个字节依次是回车符（0x0D）、换行符（0x0A）、替换符（0x1A）和另一个换行符，写成Go语言字面量的话，结果如下所示。 "\x19\x93\r\n\x1a\n"
	cintSize        byte
	sizetSize       byte
	instructionSize byte
	luaIntegerSize  byte
	luaNumberSize   byte
	luacInt         int64
	luacNum         float64
}

// 寄存器数量。这个字段也被叫作MaxStackSize，为什么这样叫呢？这是因为Lua虚拟机在执行函数时，真正使用的其实是一种栈结构，这种栈结构除了可以进行常规地推入和弹出操作以外，还可以按索引访问，所以可以用来模拟寄存器。
type Prototype struct {
	Source          string        // 源文件名
	LineDefined     uint32        // 起始行号
	LastLineDefined uint32        // 终止行号
	NumParams       byte          // 固定参数个数
	IsVararg        byte          // 是否是变长参数函数
	MaxStackSize    byte          // 寄存器数量。Lua编译器会为每一个Lua函数生成一个指令表，也就是我们常说的字节码。由于Lua虚拟机是基于寄存器的虚拟机（详见第3章），大部分指令也都会涉及虚拟寄存器操作，那么一个函数在执行期间至少需要用到多少个虚拟寄存器呢？Lua编译器会在编译函数时将这个数量计算好，并以字节类型保存在函数原型里。运行“Hello，World！”程序需要2个虚拟寄存器
	Code            []uint32      // 指令表
	Constants       []interface{} // 常量表 常量表用于存放Lua代码里出现的字面量，包括nil、布尔值、整数、浮点数和字符串五种
	Upvalues        []Upvalue     // upvalue表
	Protos          []*Prototype  // 子函数原型表
	LineInfo        []uint32      // 行号表
	LocVars         []LocVar      // 局部变量表
	UpvalueNames    []string      // upvalue名列表
}

type Upvalue struct {
	Instack byte
	Idx     byte
}

type LocVar struct {
	VarName string
	StartPC uint32
	EndPC   uint32
}

func Undump(data []byte) *Prototype {
	reader := &reader{data}
	reader.checkHeader()        // 检查二进制chunk头部
	reader.readByte()           // 跳过Upvalue数量
	return reader.readProto("") // 读取主函数原型
}

func IsBinaryChunk(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == LUA_SIGNATURE
}
