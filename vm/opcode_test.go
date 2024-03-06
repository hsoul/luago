package vm_test

import (
	"fmt"
	"luago/binchunk"
	"luago/state"
	. "luago/vm"
	"os"
	"testing"
)

func CodeTest(t *testing.T) {
	data, err := os.ReadFile("test.lua")
	if err != nil {
		panic(err)
	}
	proto := binchunk.Undump(data)
	luaMain(proto)
}

func luaMain(proto *binchunk.Prototype) {
	nRegs := int(proto.MaxStackSize)
	ls := state.New()
	ls.SetTop(nRegs)
	for {
		pc := ls.PC()
		inst := Instruction(ls.Fetch())
		if inst.Opcode() != OP_RETURN {
			inst.Execute(ls)
			fmt.Printf("[%02d] %s", pc+1, inst.OpName())
			ls.PStack()
		} else {
			break
		}
	}
}
