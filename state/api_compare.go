package state

import . "luago/api"

func (s *luaState) Compare(idx1, idx2 int, op CompareOp) bool {
	if !s.stack.isValid(idx1) || !s.stack.isValid(idx2) {
		return false
	}

	a := s.stack.get(idx1)
	b := s.stack.get(idx2)
	switch op {
	case LUA_OPEQ:
		return _eq(a, b, s)
	case LUA_OPLT:
		return _lt(a, b, s)
	case LUA_OPLE:
		return _le(a, b, s)
	default:
		panic("invalid compare op!")
	}
}

func _eq(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case int64:
		switch y := b.(type) {
		case int64:
			return x == y
		case float64:
			return float64(x) == y
		default:
			return false
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return x == float64(y)
		case float64:
			return x == y
		default:
			return false
		}
	case string:
		y, ok := b.(string)
		return ok && x == y
	case *luaTable:
		if y, ok := b.(*luaTable); ok && x != y && ls != nil { // 对于等于（==）运算，当且仅当两个操作数是不同的表时，才会尝试执行__eq元方法。
			if result, ok := callMetamethod(x, y, "__eq", ls); ok {
				return convertToBoolean(result)
			}
		}
		return a == b
	default:
		return a == b
	}
}

func _lt(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case int64:
		switch y := b.(type) {
		case int64:
			return x < y
		case float64:
			return float64(x) < y
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return x < float64(y)
		case float64:
			return x < y
		}
	case string:
		y, ok := b.(string)
		return ok && x < y
	}
	if result, ok := callMetamethod(a, b, "__lt", ls); ok {
		return convertToBoolean(result)
	} else {
		panic("comparison error!")
	}
}

func _le(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64:
		switch y := b.(type) {
		case int64:
			return x <= float64(y)
		case float64:
			return x <= y
		}
	}

	if result, ok := callMetamethod(a, b, "__le", ls); ok {
		return convertToBoolean(result)
	} else if result, ok := callMetamethod(b, a, "__lt", ls); ok {
		return !convertToBoolean(result)
	} else {
		panic("comparison error!")
	}
}

func (s *luaState) RawEqual(idx1, idx2 int) bool {
	if !s.stack.isValid(idx1) || !s.stack.isValid(idx2) {
		return false
	}
	a := s.stack.get(idx1)
	b := s.stack.get(idx2)
	return _eq(a, b, nil)
}
