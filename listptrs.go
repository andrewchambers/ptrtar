package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"os"
)

func ListPtrsUsage() {
	fmt.Println("print all ptrs to stdout")
	flag.PrintDefaults()
	os.Exit(1)
}

func printPtrs(in io.Reader, out io.Writer, newline bool) error {
	inTar := tar.NewReader(os.Stdin)
	for {
		h, err := inTar.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if h.PAXRecords != nil {
			_, ok := h.PAXRecords["PTRTAR.sz"]
			if ok {
				_, err = io.Copy(out, inTar)
				if err != nil {
					return err
				}

				if newline {
					_, err = out.Write([]byte("\n"))
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func ListPtrsMain() {
	flag.Usage = ToTarUsage
	newline := flag.Bool("n", false, "print a new line after each pointer")
	flag.Parse()

	err := printPtrs(os.Stdin, os.Stdout, *newline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing tar stream: %s\n", err)
		os.Exit(1)
	}
}
