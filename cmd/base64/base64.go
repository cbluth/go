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

// ./base64					:: waits for manual input at stdin, encodes
// ./base64 file/path		:: encodes file
// ./base64 < file/path		:: encodes file
// ./base64 <<< "hi"		:: encodes text
// echo hi | ./base64		:: encodes text
// ./base64 -w2 file/path	:: encodes file at wrapbreak
// ./base64 -w2 < file/path	:: encodes file at wrapbreak
// ./base64 -w2 <<< "hi"	:: encodes text at wrapbreak
// echo hi | ./base64 -w2	:: encodes text at wrapbreak
// ./base64 -h				:: shows help
// ./base64 -d				:: decode with no file and no stdin, wait for manual input
// ./base64 -d file/path	:: read stdin decode base64 into plaintext
// ./base64 -d < file/path	:: read stdin decode base64 into plaintext
// ./base64 -d <<< "=="		:: read stdin decode base64 into plaintext
// echo == | ./base64 -d	:: read stdin decode base64 into plaintext
// ./base64 -w3				:: set wrapbreak at 3 (default 65)

func main() {
	err := cli()
	if err != nil {
		log.Fatalln(err)
	}
}

func cli() error {
	args := struct{
		wrap int
		help bool
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
	r := (*os.File)(nil)
	if len(p) == 0 || p[0] == `-` {
		r = os.Stdin
	}
	if len(p) > 0 && p[0] != `-` {
		err := (error)(nil)
		r, err = os.Open(os.ExpandEnv(p[0]))
		if err != nil {
			return err
		}
		defer r.Close()
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
	o := []string{}
	for i := 0 ; i < len(s) ; i += wrap {
		if i+wrap < len(s) {
			o = append(o, s[i:(i+wrap)])
		} else {
			o = append(o, s[i:])
		}
	}
	fmt.Println(strings.Join(o, "\n"))
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
