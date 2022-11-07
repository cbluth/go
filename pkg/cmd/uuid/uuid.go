package uuid

import (
	"fmt"

	"github.com/cbluth/go/pkg/uuid"
)

func ExecuteCommand() error {
	u, err := uuid.New()
	if err != nil {
		return err
	}
	fmt.Println(u.String())
	return nil
}
