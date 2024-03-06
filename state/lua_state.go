package state

import . "luago/api"

type luaState struct {
	registry *luaTable // 注册表
	stack    *luaStack
	coStatus int
	coCaller *luaState
	coChan   chan int
}

func New() *luaState {
	ls := &luaState{}

	registry := newLuaTable(8, 0)
	registry.put(LUA_RIDX_MAINTHREAD, ls)
	registry.put(LUA_RIDX_GLOBALS, newLuaTable(0, 0)) // 全局环境表

	ls.registry = registry
	ls.pushLuaStack(newLuaStack(LUA_MINSTACK, ls))

	return ls
}

func (s *luaState) isMainThread() bool {
	return s.registry.get(LUA_RIDX_MAINTHREAD) == s
}

func (s *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = s.stack // 我们使用单向链表来实现函数调用栈。这个链表的头部是栈顶，尾部是栈底。往栈顶推入一个调用帧相当于在链表头部插入一个节点，并让这个节点成为新的头部。
	s.stack = stack
}

func (s *luaState) popLuaStack() {
	stack := s.stack
	s.stack = stack.prev // 从栈顶弹出一个调用帧相当于把链表头部节点删除，并把下一个节点作为新的头部。
	stack.prev = nil
}
