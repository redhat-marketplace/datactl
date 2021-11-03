//go:build tools
// +build tools

package main

import (
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/gocode"
)

func main() {
	cwd, _ := os.Getwd()
	pkg := "."

	inst := cue.Build(load.Instances([]string{pkg}, &load.Config{
		Dir:        cwd + string(os.PathSeparator) + "schema",
		ModuleRoot: cwd + string(os.PathSeparator) + "schema",
		Module:     "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/schema",
	}))[0]
	if err := inst.Err; err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b, err := gocode.Generate(".", inst, &gocode.Config{})
	if err != nil {
		// handle error
	}

	fmt.Println(string(b))
	//err = ioutil.WriteFile("cue_gen.go", b, 0644)
}
