package wasm

import (
	_ "embed"
)

//go:embed dist/sandbox.wasm
var Embedded []byte
