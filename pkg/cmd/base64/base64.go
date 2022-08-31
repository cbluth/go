package base64

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cbluth/go/pkg/base64"
)

func ExecuteCommand() error {
	args := struct {
		wrap   int
		help   bool
		decode bool
	}{}

	flag.BoolVar(&args.help, "h", false, "show help dialog")
	flag.BoolVar(&args.decode, "d", false, "decode base64 input")
	flag.IntVar(&args.wrap, "w", 65, "wrap encoding to w characters")
	flag.Usage = func() {
		fmt.Println("base64 [OPTION] [filepath]")
		flag.PrintDefaults()
	}
	flag.Parse()
	if args.help {
		flag.Usage()
		return nil
	}
	p := flag.Args()
	r := (io.Reader)(nil)
	if len(p) == 0 || p[0] == `-` {
		r = os.Stdin
	}
	err := (error)(nil)
	if len(p) > 0 && p[0] != `-` {
		r, err = os.Open(os.ExpandEnv(p[0]))
		if err != nil {
			return err
		}
	}
	out := ""
	if args.decode {
		out, err = base64.DecodeReaderString(r)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	}
	out, err = base64.EncodeReaderString(r, args.wrap)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
