//go:build pico
// +build pico

package main

import (
    "machine"
)

func init() {
    ledPin       = machine.LED
    resetBtnPin  = machine.GP18
    setBtnPin    = machine.GP19
    centerBtnPin = machine.GP20
    rightBtnPin  = machine.GP21
    leftBtnPin   = machine.GP22
    downBtnPin   = machine.GP26
    upBtnPin     = machine.GP27
}
