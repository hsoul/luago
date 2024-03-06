package api

type LuaVM interface {
	LuaState
	PC() int            // 返回当前PC
	AddPC(n int)        // 修改PC
	Fetch() uint32      // 获取当前指令，PC++
	GetConst(idx int)   // 将指定常量推入栈顶
	GetRK(rk int)       // 将指定常量或栈值推入栈顶
	RegisterCount() int // 返回当前函数注册数量
	LoadVararg(n int)   // 加载函数的变长参数到栈顶
	LoadProto(idx int)  // 加载子函数原型到栈顶
	CloseUpvalues(a int)
}
