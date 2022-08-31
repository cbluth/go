package base64

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"strings"
)

func EncodeReaderString(r io.Reader, wrap int) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
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
	return strings.Join(w, "\n"), nil
}

func DecodeReaderString(r io.Reader) (string, error) {
	e, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	b, err := base64.StdEncoding.DecodeString(string(e))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
