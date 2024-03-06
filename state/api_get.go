package state

import . "luago/api"

func (s *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	s.stack.push(t)
}

func (s *luaState) NewTable() {
	s.CreateTable(0, 0)
}

func (s *luaState) GetTable(idx int) LuaType {
	t := s.stack.get(idx)
	k := s.stack.pop()
	return s.getTable(t, k, false)
}

// push(t[k])
func (s *luaState) getTable(t, k luaValue, raw bool) LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			s.stack.push(v)
			return typeOf(v)
		}
	}

	if !raw {
		if mf := getMetafield(t, "__index", s); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return s.getTable(x, k, false)
			case *closure:
				s.stack.push(mf)
				s.stack.push(t)
				s.stack.push(k)
				s.Call(2, 1)
				v := s.stack.get(-1)
				return typeOf(v)
			}
		}
	}

	panic("index error!")
}

func (s *luaState) GetField(idx int, k string) LuaType {
	t := s.stack.get(idx)
	return s.getTable(t, k, false)
}

func (s *luaState) GetI(idx int, i int64) LuaType {
	t := s.stack.get(idx)
	return s.getTable(t, i, false)
}

func (s *luaState) GetGlobal(name string) LuaType {
	t := s.registry.get(LUA_RIDX_GLOBALS)
	return s.getTable(t, name, false)
}

func (s *luaState) RawGet(idx int) LuaType {
	t := s.stack.get(idx)
	k := s.stack.pop()
	return s.getTable(t, k, true)
}

func (s *luaState) RawGetI(idx int, i int64) LuaType {
	t := s.stack.get(idx)
	return s.getTable(t, i, true)
}

func (s *luaState) GetMetatable(idx int) bool {
	val := s.stack.get(idx)
	if mt := getMetatable(val, s); mt != nil {
		s.stack.push(mt)
		return true
	} else {
		return false
	}
}
