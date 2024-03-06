package state

import (
	. "luago/api"
	"testing"
)

// 编写测试代码和测试命令
// go test -v -run TestAdd
func TestArith(t *testing.T) {
	ls := New()
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)
	PrintStack(ls)

	ls.Arith(LUA_OPADD)
	PrintStack(ls)

	ls.Arith(LUA_OPBNOT)
	PrintStack(ls)

	ls.Len(2)
	PrintStack(ls)

	ls.Concat(3)
	PrintStack(ls)

	ls.PushBoolean(ls.Compare(1, 2, LUA_OPEQ))
	PrintStack(ls)

	// fmt.Println(^7)
	// fmt.Println(reflect.TypeOf(ls.stack.get(1)))
}
