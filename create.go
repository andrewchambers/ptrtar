package main

import (
	"archive/tar"
	"bytes"
	"container/list"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type ptrTarWorkItem struct {
	AbsPath    string
	PassedPath string
	Stat       os.FileInfo
}

func runPtrFunc(fpath string, ptrFunc func(io.Reader, io.Writer) error, out io.Writer) (int64, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	metered := MeteredReader{R: f}

	err = ptrFunc(&metered, out)
	if err != nil {
		return 0, err
	}

	return metered.ReadCount, nil
}

func hostToPtrTar(cache *CreateCache, workList *list.List, exclude map[string]struct{}, ptrFunc func(io.Reader, io.Writer) error, tarWriter *tar.Writer) error {

	for workList.Len() != 0 {

		workItem := workList.Remove(workList.Front()).(ptrTarWorkItem)
		workItemSysStat := workItem.Stat.Sys().(*syscall.Stat_t)
		workItemCTime := time.Unix(int64(workItemSysStat.Ctim.Sec), int64(workItemSysStat.Ctim.Nsec))

		_, doExclude := exclude[workItem.AbsPath]
		if doExclude {
			continue
		}

		linkTarget := ""
		if (workItem.Stat.Mode() & os.ModeSymlink) != 0 {
			target, err := os.Readlink(workItem.AbsPath)
			if err != nil {
				return err
			}
			linkTarget = target
		}

		tarHeader, err := tar.FileInfoHeader(workItem.Stat, linkTarget)
		if err != nil {
			return err
		}

		tarHeader.Name = workItem.PassedPath

		var ptrBuffer bytes.Buffer
		isPtr := false

		switch {
		case workItem.Stat.Mode().IsRegular():
			isPtr = true
			tarHeader.PAXRecords = make(map[string]string)

			cacheHit := false

			if cache != nil {
				cachedPtr, ok, err := cache.HasPtr(workItem.AbsPath, workItem.Stat.ModTime(), workItemCTime, workItem.Stat.Size())
				if err != nil {
					return err
				}

				if ok {
					cacheHit = true
					_, err = ptrBuffer.Write(cachedPtr)
					if err != nil {
						return err
					}
					tarHeader.PAXRecords["PTRTAR.sz"] = fmt.Sprintf("%d", workItem.Stat.Size())
				}
			}

			if !cacheHit {
				// XXX, perhaps if the pointer starts getting huge, buffer
				// it in a /tmp/ file, otherwise just reject large pointers
				// they don't make sense to me, but maybe they have a use.

				nCopied, err := runPtrFunc(workItem.AbsPath, ptrFunc, &ptrBuffer)
				if err != nil {
					return err
				}
				tarHeader.PAXRecords["PTRTAR.sz"] = fmt.Sprintf("%d", nCopied)

				if cache != nil {
					err = cache.AddPtr(workItem.AbsPath, workItem.Stat.ModTime(), workItemCTime, nCopied, ptrBuffer.Bytes())
					if err != nil {
						return err
					}
				}
			}

		case workItem.Stat.IsDir():
			dir, err := ioutil.ReadDir(workItem.AbsPath)
			if err != nil {
				return err
			}

			for _, st := range dir {
				workList.PushFront(ptrTarWorkItem{
					PassedPath: filepath.Join(workItem.PassedPath, st.Name()),
					AbsPath:    filepath.Join(workItem.AbsPath, st.Name()),
					Stat:       st,
				})
			}
		}

		if isPtr {
			tarHeader.Size = int64(ptrBuffer.Len())
		}

		err = tarWriter.WriteHeader(tarHeader)
		if err != nil {
			return err
		}

		if isPtr {
			_, err = io.Copy(tarWriter, &ptrBuffer)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func HostToPtrTar(cache *CreateCache, paths []string, exclude map[string]struct{}, ptrFunc func(io.Reader, io.Writer) error, out io.Writer) error {

	workList := list.New()

	for _, path := range paths {

		st, err := os.Stat(path)
		if err != nil {
			return err
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		workList.PushFront(ptrTarWorkItem{
			AbsPath:    abs,
			PassedPath: path,
			Stat:       st,
		})
	}

	absExclude := make(map[string]struct{})
	for ex := range exclude {
		abs, err := filepath.Abs(ex)
		if err != nil {
			return err
		}
		absExclude[abs] = struct{}{}
	}

	tarWriter := tar.NewWriter(out)

	err := hostToPtrTar(cache, workList, absExclude, ptrFunc, tarWriter)
	if err != nil {
		return err
	}

	return tarWriter.Close()
}

func CreateUsage() {
	fmt.Println("TODO")
	flag.PrintDefaults()
	os.Exit(1)
}

type stringSet map[string]struct{}

func (e *stringSet) String() string {
	return fmt.Sprintf("%v", map[string]struct{}(*e))
}

func (e *stringSet) Set(value string) error {
	(*e)[value] = struct{}{}
	return nil
}

func CreateMain() {
	flag.Usage = CreateUsage
	exclude := make(stringSet)
	dirs := make(stringSet)
	cachePath := flag.String("cache", "", "path to write cache, (cache invalidation via rm is up to you!)")
	flag.Var(&exclude, "exclude", "paths to exclude from archive, can be specified multiple times")
	flag.Var(&dirs, "dir", "dir to archive, can be specified multiple times")

	flag.Parse()

	cmdArgs := flag.Args()
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "You didn't specify a command...\n")
		CreateUsage()
	}

	getPtr := func(in io.Reader, out io.Writer) error {
		var cmd *exec.Cmd
		if len(cmdArgs) == 1 {
			cmd = exec.Command(cmdArgs[0])
		} else {
			cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
		}

		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		cmd.Stdin = in

		return cmd.Run()
	}

	var dedupedDirs []string
	for d, _ := range dirs {
		dedupedDirs = append(dedupedDirs, d)
	}

	var cache *CreateCache
	var err error

	if *cachePath != "" {
		exclude.Set(*cachePath)
		cache, err = OpenCache(*cachePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "non fatal error opening cache: %s\n", err)
			cache = nil
		}
	}

	err = HostToPtrTar(cache, dedupedDirs, exclude, getPtr, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during archiving: %s\n", err)
		os.Exit(1)
	}

	if cache != nil {
		err = cache.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "non fatal error closing cache: %s\n", err)
		}
	}
}
