package state

import (
	. "luago/api"
	"luago/number"
	"math"
)

var (
	iadd  = func(a, b int64) int64 { return a + b }
	fadd  = func(a, b float64) float64 { return a + b }
	isub  = func(a, b int64) int64 { return a - b }
	fsub  = func(a, b float64) float64 { return a - b }
	imul  = func(a, b int64) int64 { return a * b }
	fmul  = func(a, b float64) float64 { return a * b }
	imod  = number.IMod
	fmod  = number.FMod
	pow   = math.Pow
	div   = func(a, b float64) float64 { return a / b }
	iidiv = number.IFloorDiv
	fidiv = number.FFloorDiv
	band  = func(a, b int64) int64 { return a & b }
	bor   = func(a, b int64) int64 { return a | b }
	bxor  = func(a, b int64) int64 { return a ^ b }
	shl   = number.ShiftLeft
	shr   = number.ShiftRight
	iunm  = func(a, _ int64) int64 { return -a } // _ 表示不使用第二个参数
	funm  = func(a, _ float64) float64 { return -a }
	bnot  = func(a, _ int64) int64 { return ^a } // ^ 按位取反
)

type operator struct {
	metamethod  string
	integerFunc func(int64, int64) int64
	floatFunc   func(float64, float64) float64
}

var operators = []operator{
	operator{"__add", iadd, fadd},    // LUA_OPADD
	operator{"__sub", isub, fsub},    // LUA_OPSUB
	operator{"__mul", imul, fmul},    // LUA_OPMUL
	operator{"__mod", imod, fmod},    // LUA_OPMOD
	operator{"__pow", nil, pow},      // LUA_OPPOW
	operator{"__div", nil, div},      // LUA_OPDIV
	operator{"__idiv", iidiv, fidiv}, // LUA_OPIDIV
	operator{"__band", band, nil},    // LUA_OPBAND
	operator{"__bor", bor, nil},      // LUA_OPBOR
	operator{"__bxor", bxor, nil},    // LUA_OPBXOR
	operator{"__shl", shl, nil},      // LUA_OPSHL
	operator{"__shr", shr, nil},      // LUA_OPSHR
	operator{"__umn", iunm, funm},    // LUA_OPUNM
	operator{"__bnot", bnot, nil},    // LUA_OPBNOT
}

func (s *luaState) Arith(op ArithOp) {
	var a, b luaValue
	b = s.stack.pop()
	if op != LUA_OPUNM && op != LUA_OPBNOT { // 二元运算符
		a = s.stack.pop()
	} else { // 一元运算符
		a = b
	}

	operator := operators[op]
	if result := _arith(a, b, operator); result != nil {
		s.stack.push(result)
		return
	}

	mm := operator.metamethod
	if result, ok := callMetamethod(a, b, mm, s); ok {
		s.stack.push(result)
		return
	}

	panic("arithmetic error!")
}

func _arith(a, b luaValue, op operator) luaValue {
	if op.floatFunc == nil {
		if x, ok := convertToInteger(a); ok {
			if y, ok := convertToInteger(b); ok {
				return op.integerFunc(x, y)
			}
		}
	} else {
		if op.integerFunc != nil { // add sub mul mod idiv unm
			if x, ok := a.(int64); ok {
				if y, ok := b.(int64); ok {
					return op.integerFunc(x, y)
				}
			}
		}

		if x, ok := convertToFloat(a); ok {
			if y, ok := convertToFloat(b); ok {
				return op.floatFunc(x, y)
			}
		}
	}
	return nil
}
