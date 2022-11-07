package cat

import (
	"io"
	"os"
)

func ExecuteCommand() error {
	for _, path := range os.Args[1:] {
		f, err := os.Open(os.ExpandEnv(path))
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(os.Stdout, f)
		if err != nil {
			return err
		}
	}
	return nil
}
