package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func ToTarUsage() {
	fmt.Println("convert a ptrtar on stdin to a normal tar on stdout,")
	fmt.Println("using the provided program to fetch pointer data.")
	flag.PrintDefaults()
	os.Exit(1)
}

func toTar(in io.Reader, out io.Writer, getData func(inPtr io.Reader, out io.Writer) error) error {
	inTar := tar.NewReader(os.Stdin)
	outTar := tar.NewWriter(os.Stdout)

	for {
		h, err := inTar.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var trueSz int64
		var isPtr bool

		if h.PAXRecords != nil {
			trueSzStr, ok := h.PAXRecords["PTRTAR.sz"]
			if ok {
				trueSz, _ = strconv.ParseInt(trueSzStr, 10, 64)
				isPtr = true
			}
		}

		if isPtr {
			delete(h.PAXRecords, "PTRTAR.sz")
			h.Size = trueSz
			err = outTar.WriteHeader(h)
			if err != nil {
				return err
			}

			metered := MeteredWriter{W: outTar}

			err := getData(inTar, &metered)
			if err != nil {
				return err
			}
			if metered.WriteCount != trueSz {
				return fmt.Errorf("file %s, was %d bytes but header expected %d bytes", h.Name, metered.WriteCount, trueSz)
			}

		} else {
			err = outTar.WriteHeader(h)
			if err != nil {
				return err
			}

			_, err = io.Copy(outTar, inTar)
			if err != nil {
				return err
			}
		}
	}

	err := outTar.Close()
	if err != nil {
		return err
	}

	return nil
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

	getData := func(inPtr io.Reader, out io.Writer) error {

		var cmd *exec.Cmd
		if len(cmdArgs) == 1 {
			cmd = exec.Command(cmdArgs[0])
		} else {
			cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
		}

		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		cmd.Stdin = inPtr

		return cmd.Run()
	}

	err := toTar(os.Stdin, os.Stdout, getData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing tar stream: %s\n", err)
		os.Exit(1)
	}
}
