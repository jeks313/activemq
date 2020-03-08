//+build mage

package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Runs dep ensure and then installs the binary.
func Build() error {
	mg.Deps(Pkger)
	return sh.Run("go", "build")
}

func Pkger() error {
	if err := sh.Run("pkger"); err != nil {
		if strings.Index(err.Error(), "executable file not found") < 0 {
			return err
		}
		fmt.Println("getting pkger...")
		return sh.Run("go", "get", "github.com/markbates/pkger/cmd/pkger")
	}
	return nil
}
