package main

import (
    "fmt"
    "machine"
    "time"

    "github.com/elehobica/pico_tinygo_timer_alarm/mymachine"
)

var (
    ledPin   machine.Pin
    serial  = machine.Serial
)

type Pin struct {
    *machine.Pin
}

var led *Pin = &Pin{&ledPin}

func (pin Pin) Toggle() {
    pin.Set(!pin.Get())
}

func main() {
    println(); println()
    println("=============================")
    println("== pico_tinygo_timer_alarm ==")
    println("=============================")

    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    led.Low()

    mymachine.SetAlarmInterrupt(5*1000*1000, alarm)

    for {
        time.Sleep(1000 * time.Millisecond)
        fmt.Printf("loop\r\n");
    }
}

func alarm() {
    fmt.Printf("alarm\r\n");
    led.Toggle()
}
