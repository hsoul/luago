package api

type LuaType = int
type ArithOp = int   // 算术运算符
type CompareOp = int // 比较运算符

type GoFunction func(LuaState) int

func LuaUpvalueIndex(i int) int {
	return LUA_REGISTRYINDEX - i
}

type LuaState interface {
	BasicAPI // 基础API
	AuxLib   // 辅助API
}

type BasicAPI interface {
	// basic stack manipulation
	GetTop() int             // 方法返回栈顶索引
	AbsIndex(idx int) int    // 方法将索引 idx 转换为绝对索引
	CheckStack(n int) bool   // Lua栈的容量并不会自动增长，API使用者必须在必要的时候调用CheckStack（）方法检查栈剩余空间，看是否还可以推入n个值而不会导致溢出。如果剩余空间足够或者扩容成功则返回true，否则返回false。以图4-5为例（左边是执行操作前栈的状态，右边是执行操作后栈的状态），栈里已经推入3个值，还有1个剩余空间。执行CheckStack（3）之后，栈应该有至少3个剩余空间。
	Pop(n int)               // 从栈顶弹出n个值
	Copy(fromIdx, toIdx int) // 把值从一个位置复制到另一个位置
	PushValue(idx int)       // 把指定索引处的值推入栈顶
	Replace(idx int)         // PushValue（）的反操作：将栈顶值弹出，然后写入指定位置
	Insert(idx int)          // Insert（）方法将栈顶值弹出，然后插入指定位置。以图4-10为例，栈里有5个元素，执行Insert（2）之后，栈顶值被弹出并插入到索引2处，原来位于索引2、3、4处的值则分别上移了一个位置。
	Remove(idx int)          // 删除指定索引处的值，然后将该值上面的值全部下移一个位置
	Rotate(idx, n int)       // 将[idx，top]索引区间内的值朝栈顶方向旋转n个位置。如果n是负数，那么实际效果就是朝栈底方向旋转。
	SetTop(idx int)          // 将栈顶索引设置为指定值。如果指定值小于当前栈顶索引，效果则相当于弹出操作（指定值为0相当于清空栈）。以图4-14为例，原栈顶索引是5，执行SetTop（3）之后，栈顶2个值被弹出，栈顶索引变为3。如果指定值大于当前栈顶索引，则效果相当于推入多个nil值。
	// access functions (stack -> Go)
	TypeName(tp LuaType) string
	Type(idx int) LuaType
	IsNone(idx int) bool
	IsNil(idx int) bool
	IsNoneOrNil(idx int) bool
	IsBoolean(idx int) bool
	IsInteger(idx int) bool
	IsNumber(idx int) bool
	IsString(idx int) bool
	IsTable(idx int) bool
	IsThread(idx int) bool
	IsFunction(idx int) bool
	IsGoFunction(idx int) bool // 判断给定索引处的值是否是Go函数
	ToBoolean(idx int) bool
	ToInteger(idx int) int64
	ToIntegerX(idx int) (int64, bool)
	ToNumber(idx int) float64
	ToNumberX(idx int) (float64, bool)
	ToString(idx int) string
	ToStringX(idx int) (string, bool)
	ToGoFunction(idx int) GoFunction
	ToPointer(idx int) interface{}
	RawLen(idx int) uint
	// push functions (Go -> stack)
	PushNil()
	PushBoolean(b bool)
	PushInteger(n int64)
	PushNumber(n float64)
	PushString(s string)
	PushFString(fmt string, a ...interface{})
	PushGoFunction(f GoFunction)       // 将Go函数推入栈顶
	PushGoClosure(f GoFunction, n int) // 创建一个Go闭包，将其推入栈顶
	PushGlobalTable()                  // 将全局环境表推入栈顶
	// comparison and arithmetic functions
	Arith(op ArithOp)
	Compare(idx1, idx2 int, op CompareOp) bool
	RawEqual(idx1, idx2 int) bool
	// get functions (Lua -> stack)
	NewTable()                  // 创建一个空表，将其推入栈顶
	CreateTable(nArr, nRec int) // 创建一个空表，预分配nArr个数组元素和nRec个非数组元素，将其推入栈顶
	GetTable(idx int) LuaType   // 从栈顶弹出一个键，然后以该键为索引从指定位置的table中取出一个值，最后将该值推入栈顶
	GetField(idx int, k string) LuaType
	GetI(idx int, i int64) LuaType
	RawGet(idx int) LuaType
	RawGetI(idx int, i int64) LuaType
	GetMetatable(idx int) bool
	GetGlobal(name string) LuaType // 从全局环境表中获取一个字段的值，然后推入栈顶

	// set functions (Go -> stack)
	SetTable(idx int)           // 从栈顶依次弹出value、key, 然后把value赋给指定索引处table[key]
	SetField(idx int, k string) // 从栈顶弹出一个值，然后把该值赋给指定索引处table[k]
	SetI(idx int, i int64)
	RawSet(idx int)
	RawSetI(idx int, i int64)
	SetMetatable(idx int)
	SetGlobal(name string)              // 从栈顶弹出一个值，然后将其设为全局环境表的一个字段
	Register(name string, f GoFunction) // 将Go函数注册到全局环境表

	// 'load' and 'call' functions (load and run Lua code)
	Load(chunk []byte, chunkName, mode string) int // Load（）方法加载二进制chunk，把主函数原型实例化为闭包并推入栈顶。实际上该方法不仅可以加载预编译的二进制chunk，也可以直接加载Lua脚本。如果加载的是二进制chunk，那么只要读取文件、解析主函数原型、实例化为闭包、推入栈顶就可以了；如果加载的是Lua脚本，则要先进行编译。为了简化描述，后面把二进制chunk和Lua脚本统称为chunk。
	Call(nArgs, nResults int)                      // Call（）方法对Lua函数进行调用。在执行Call（）方法之前，必须先把被调函数推入栈顶，然后把参数值依次推入栈顶。Call（）方法结束之后，参数值和函数会被弹出栈顶，取而代之的是指定数量的返回值。Call（）方法接收两个参数：第一个参数指定准备传递给被调函数的参数数量，同时也隐含给出了被调函数在栈里的位置；第二个参数指定需要的返回值数量（多退少补），如果是-1，则被调函数的返回值会全部留在栈顶。
	PCall(nArgs, nResults, msgh int) int
	/* miscellaneous functions */
	Len(idx int)  // 访问指定索引处的值，取其长度，然后推入栈顶
	Concat(n int) // 从栈顶弹出n个值，对这些值进行拼接，然后把结果推入栈顶
	Next(idx int) bool
	Error() int
	StringToNumber(s string) bool

	// corioutine functions
	NewThread() LuaState
	Resume(from LuaState, nArgs int) int
	Yield(nResults int) int
	Status() int
	IsYieldable() bool
	ToThread(idx int) LuaState
	PushThread() bool
	XMove(to LuaState, n int)
	GetStack() bool

	// debug
	PStack()
}
