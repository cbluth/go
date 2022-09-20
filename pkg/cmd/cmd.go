package cmd

import (
	"github.com/cbluth/go/pkg/cmd/base64"
	"github.com/cbluth/go/pkg/cmd/cat"
	"github.com/cbluth/go/pkg/cmd/uuid"
)

func Base64() error {
	return base64.ExecuteCommand()
}

func Cat() error {
	return cat.ExecuteCommand()
}

func UUID() error {
	return uuid.ExecuteCommand()
}
