// Package main provides the MemoNexus Go Core library entry point.
// This is a platform-agnostic library that can be compiled as:
// - Shared library for mobile (Dart FFI)
// - Standalone binary for desktop
package main

import (
	"fmt"
	"log"
)

// Version is set at build time
var Version = "0.1.0"

func main() {
	fmt.Printf("MemoNexus Core v%s\n", Version)
	log.Println("MemoNexus Go Core - Platform-Agnostic Library")
}
