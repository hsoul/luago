调用函数流程 = {
    通过指令依次将被调用函数及参数加载到寄存器
    调用 inst.call = {
        将寄存器中的函数和参数依次压入栈
        调用vm.Call = {
            创建新的调用栈（callee）
            将caller栈中的参数传递给callee栈
            将callee栈推入调用栈
            执行被调用函数指令
            将callee栈弹出调用栈
            将callee栈中的返回值传递给caller栈
        }
        将栈中的返回值依次弹出到寄存器
    }
}
