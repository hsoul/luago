package state

import "testing"

// 编写测试代码和测试命令
// go test -v -run TestAdd
func TestApi(t *testing.T) {
	ls := New()
	ls.PushBoolean(true)
	PrintStack(ls)
	ls.PushInteger(10)
	PrintStack(ls)
	ls.PushNil()
	PrintStack(ls)
	ls.PushString("hello")
	PrintStack(ls)
	ls.PushValue(-4)
	PrintStack(ls)
	ls.Replace(3)
	PrintStack(ls)
	ls.SetTop(6)
	PrintStack(ls)
	ls.Remove(-3)
	PrintStack(ls)
	ls.SetTop(-5)
	PrintStack(ls)
}
