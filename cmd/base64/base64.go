package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	err := cli()
	if err != nil {
		log.Fatalln(err)
	}
}

func cli() error {
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
	if len(p) > 0 && p[0] != `-` {
		err := (error)(nil)
		r, err = os.Open(os.ExpandEnv(p[0]))
		if err != nil {
			return err
		}
	}
	if args.decode {
		return decode(r)
	}
	return encode(r, args.wrap)
}

func encode(r io.Reader, wrap int) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s := base64.StdEncoding.EncodeToString(b)
	w := []string{}
	for i := 0; i < len(s); i += wrap {
		if i+wrap < len(s) {
			w = append(w, s[i:(i+wrap)])
		} else {
			w = append(w, s[i:])
		}
	}
	fmt.Println(strings.Join(w, "\n"))
	return nil
}

func decode(r io.Reader) error {
	e, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	b, err := base64.StdEncoding.DecodeString(string(e))
	if err != nil {
		return err
	}
	_, err = fmt.Print(string(b))
	return err
}
