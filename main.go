package main

import (
	"fmt"
	"os"
)

func Usage() {
	fmt.Println("please specify one of the subcommands:")
	fmt.Println("(c)reate, e(x)tract, list-ptrs")
	os.Exit(1)
}

func main() {

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "c", "create":
		os.Args = os.Args[1:]
		CreateMain()
	default:
		Usage()
	}
}
