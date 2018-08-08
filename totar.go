package main

import (
	"flag"
	"fmt"
	"os"
	//"io"
)

func ToTarUsage() {
	fmt.Println("TODO")
	flag.PrintDefaults()
	os.Exit(1)
}

func ToTarMain() {
	flag.Usage = ToTarUsage

	flag.Parse()

	cmdArgs := flag.Args()
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "You didn't specify a command...\n")
		ToTarUsage()
		return
	}

	/*
		getData := func(inPtr io.Reader, out io.Writer) error {

			var cmd *exec.Cmd
			if len(cmdArgs) > 1 {
				cmd = exec.Command(cmdArgs[0])
			} else {
				cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
			}

			f, err := os.Open(fpath)
			if err != nil {
				return err
			}
			defer f.Close()

			cmd.Stdout = out
			cmd.Stderr = os.Stderr
			cmd.Stdin = f

			return cmd.Run()

			return nil
		}
	*/

	/*
		err = PtrTarToTar(cache, dedupedDirs, exclude, getPtr, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error during archiving: %s\n", err)
			os.Exit(1)
		}
	*/
}
