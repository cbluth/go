package jlog

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func Println(v map[string]any) {
	v["timestamp"] = time.Now().UTC().UnixNano()
	j, err := json.Marshal(v)
	if err == nil {
		fmt.Fprintln(os.Stderr, string(j))
	}
}
