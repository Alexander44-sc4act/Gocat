//go:build !unix

package cmd

import "os"

func setupStabilizeResize(ptmx *os.File) func() {
	return func() {}
}
