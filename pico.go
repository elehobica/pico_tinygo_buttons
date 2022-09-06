//go:build pico
// +build pico

package main

import (
    "machine"
)

func init() {
    ledPin   = machine.LED
}
