package main

import (
    "fmt"
    "machine"
    "time"
    "os"

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
        if (loop == 10) {
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM0, 0.1*1000*1000, alarm0)
            mymachine.SetOneshotTimerAlarm(mymachine.ALARM1, 5*1000*1000, alarm1)
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM2, 3*1000*1000, alarm2)
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM3, 1*1000*1000, alarm3)
            fmt.Printf("setting changed 1\r\n");
        } else if (loop == 20) {
            mymachine.SetRepeatedTimerAlarm(mymachine.ALARM1, 0, alarm1)
            fmt.Printf("setting changed 2\r\n");
        } else if (loop == 30) {
            // to check if ALARM3 fires in spite of too busy ALARM1
            mymachine.SetOneshotTimerAlarm(mymachine.ALARM3, 3*1000*1000, exit)
            fmt.Printf("setting changed 3\r\n");
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

func exit() {
    fmt.Printf("exit\r\n");
    os.Exit(0)
}