package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

var vmPool = sync.Pool{
	New: func() interface{} {
		L := lua.NewState()
		L.PreloadModule("fs", fsLoader)
		L.PreloadModule("router", routerLoader)
		L.PreloadModule("http", httpLoader)
		L.PreloadModule("json", jsonLoader)
		return L
	},
}

func main() {
	L := vmPool.Get().(*lua.LState)

	defer func() {
		L.Close()
		vmPool.Put(L)
	}()

	args := parseArgs()

	if _, ok := args["v"]; ok {
		fmt.Println(ReleasesName)
		os.Exit(0)
	}

	if args["file"] == "" {
		args["file"] = filepath.Join(getExeDir(), "main.lua")
	}
	mainLua, err := resolveFile(args["file"])
	if err != nil {
		log.Fatalf("File not found: %s", mainLua)
	}
	mainDir := filepath.Dir(mainLua)

	packagePath := `package.path = package.path .. ';` + mainDir + `/?.lua'`
	if err := L.DoString(packagePath); err != nil {
		log.Fatalf("Failed to set package.path: %v", err)
	}

	argsTable := L.NewTable()
	for key, arg := range args {
		argsTable.RawSet(lua.LString(key), lua.LString(arg))
	}
	L.SetGlobal("arg", argsTable)

	if err := L.DoFile(mainLua); err != nil {
		panic(err)
	}
}
