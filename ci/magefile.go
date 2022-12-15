//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/spf13/viper/ci/lib"
)

// Run tests
func Test(ctx context.Context) error {
	fmt.Println("Running tests...")

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	path, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	output, err := lib.Test(client, path, true).Stdout(ctx)
	if err != nil {
		return err
	}

	fmt.Print(output)

	return nil
}
