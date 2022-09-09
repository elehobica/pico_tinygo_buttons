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

    mymachine.SetRepeatedTimerAlarm(mymachine.ALARM0, 0.05*1000*1000, alarm0)
    mymachine.SetRepeatedTimerAlarm(mymachine.ALARM1, 1*1000*1000, alarm1)
    mymachine.SetOneshotTimerAlarm(mymachine.ALARM2, 2*1000*1000, alarm2)
    mymachine.SetRepeatedTimerAlarm(mymachine.ALARM3, 4*1000*1000, alarm3)

    for loop := 0; true; loop++ {
        time.Sleep(1000 * time.Millisecond)
        fmt.Printf("loop\r\n");
        if (loop == 20) {
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM0, 0.1*1000*1000, alarm0)
            mymachine.SetOneshotTimerAlarm(mymachine.ALARM1, 5*1000*1000, alarm1)
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM2, 3*1000*1000, alarm2)
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM3, 1*1000*1000, alarm3)
            fmt.Printf("setting changed\r\n");
        }
    }
}

func alarm0() {
    //fmt.Printf("alarm0\r\n");
    led.Toggle()
}

func alarm1() {
    fmt.Printf("alarm1\r\n");
}

func alarm2() {
    fmt.Printf("alarm2\r\n");
}

func alarm3() {
    fmt.Printf("alarm3\r\n");
}
