package main

import (
	"io/ioutil"
	. "luago/api"
	"luago/state"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		data, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}

		ls := state.New()
		ls.OpenLibs()
		// ls.LoadFile(os.Args[1])
		ls.Load(data, os.Args[1], "b")
		ls.Call(0, LUA_MULTRET)
	}
}
