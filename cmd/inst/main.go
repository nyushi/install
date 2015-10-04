package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/alecthomas/kingpin"
	"github.com/nyushi/install"
)

var (
	owner = kingpin.Flag("owner", "").Short('o').String()
	group = kingpin.Flag("group", "").Short('g').String()
	mode  = kingpin.Flag("mode", "").Short('m').String()
	dir   = kingpin.Flag("directory", "").Short('d').Bool()
	_args = kingpin.Arg("src", "").Strings()
)

func main() {
	kingpin.Parse()

	opt := &install.InstallOption{}
	if *owner != "" {
		opt.Owner = *owner
	}
	if *group != "" {
		opt.Group = *group
	}
	if *mode != "" {
		m, err := strconv.ParseInt(*mode, 8, 32)
		if err != nil {
			fmt.Printf("invalid mode: %s", err)
			os.Exit(-1)
		}
		mode := os.FileMode(m)
		opt.Mode = &mode
	}

	args := *_args
	if *dir {
		for _, name := range args {
			install.InstallDir(name, opt)
		}
		return
	}

	sources := args[:len(args)-1]
	dst := args[len(args)-1]
	if len(sources) > 1 {
		if err := checkDst(dst); err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
	}
	installFile(sources, dst, opt)
}

func installFile(sources []string, dst string, opt *install.InstallOption) {
	for _, src := range sources {
		if err := install.InstallFile(src, dst, opt); err != nil {
			fmt.Printf("error %s to %s: %s", src, dst, err)
			os.Exit(-1)
		}
	}
}
func checkDst(dst string) error {
	fi, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("error: %s: %s", dst, err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("error: dst must be directory when multiple sources")
	}
	return nil
}
