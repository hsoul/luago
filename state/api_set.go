package state

import . "luago/api"

func (s *luaState) SetTable(idx int) {
	t := s.stack.get(idx)
	v := s.stack.pop()
	k := s.stack.pop()
	s.setTable(t, k, v, false)
}

func (s *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}

	if !raw {
		if mf := getMetafield(t, "__newindex", s); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				s.setTable(x, k, v, false)
				return
			case *closure:
				s.stack.push(mf)
				s.stack.push(t)
				s.stack.push(k)
				s.stack.push(v)
				s.Call(3, 0)
				return
			}
		}
	}

	panic("not a table!")
}

func (s *luaState) SetField(idx int, k string) {
	t := s.stack.get(idx)
	v := s.stack.pop()
	s.setTable(t, k, v, false)
}

func (s *luaState) SetI(idx int, i int64) {
	t := s.stack.get(idx)
	v := s.stack.pop()
	s.setTable(t, i, v, false)
}

func (s *luaState) RawSet(idx int) {
	t := s.stack.get(idx)
	v := s.stack.pop()
	k := s.stack.pop()
	s.setTable(t, k, v, true)
}

func (s *luaState) RawSetI(idx int, i int64) {
	t := s.stack.get(idx)
	v := s.stack.pop()
	s.setTable(t, i, v, true)
}

func (s *luaState) SetGlobal(name string) {
	t := s.registry.get(LUA_RIDX_GLOBALS)
	v := s.stack.pop()
	s.setTable(t, name, v, false)
}

func (s *luaState) Register(name string, f GoFunction) {
	s.PushGoFunction(f)
	s.SetGlobal(name)
}

func (s *luaState) SetMetatable(idx int) {
	val := s.stack.get(idx)
	mtVal := s.stack.pop()

	if mtVal == nil {
		setMetatable(val, nil, s)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, s)
	} else {
		panic("table expected!")
	}
}
