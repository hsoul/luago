package state

import (
	. "luago/api"
)

func (s *luaState) NewThread() LuaState {
	t := &luaState{registry: s.registry}
	t.pushLuaStack(newLuaStack(LUA_MINSTACK, t))
	s.stack.push(t)
	return t
}

func (s *luaState) Resume(from LuaState, nArgs int) (status int) {
	lsFrom := from.(*luaState)
	if lsFrom.coChan == nil {
		lsFrom.coChan = make(chan int)
	}

	if s.coChan == nil { // start coroutine
		s.coChan = make(chan int)
		s.coCaller = lsFrom
		go func() {
			s.coStatus = s.PCall(nArgs, -1, 0)
			lsFrom.coChan <- 1
		}()
	} else { // resume coroutine
		s.coStatus = LUA_OK
		s.coChan <- 1
	}

	<-lsFrom.coChan // wait coroutine to finish or yield
	return s.coStatus
}

func (s *luaState) Yield(nResults int) int {
	s.coStatus = LUA_YIELD
	s.coCaller.coChan <- 1
	<-s.coChan
	return s.GetTop()
}

func (s *luaState) IsYieldable() bool {
	if s.isMainThread() {
		return false
	}
	return s.coStatus != LUA_YIELD
}

func (s *luaState) Status() int {
	return s.coStatus
}

func (s *luaState) GetStack() bool {
	return s.stack.prev != nil
}
