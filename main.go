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

var count [4]int

func main() {
    println(); println()
    println("=============================")
    println("== pico_tinygo_timer_alarm ==")
    println("=============================")

    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    led.Low()

    mymachine.SetRepeatedTimerAlarm("alarm0", mymachine.ALARM0, 4*1000*1000, alarm, 0, true)
    mymachine.SetRepeatedTimerAlarm("alarm1", mymachine.ALARM1, 1*1000*1000, alarm, 1, true)
    mymachine.SetRepeatedTimerAlarm("alarm2", mymachine.ALARM2, 0.05*1000*1000, ledToggle, 2)
    mymachine.SetOneshotTimerAlarm("alarm3", mymachine.ALARM3, 2*1000*1000, alarm, 3, true)

    for loop := 0; true; loop++ {
        time.Sleep(1000 * time.Millisecond)

        fmt.Printf("loop: %d", loop)
        for i := 0; i < 4; i++ {
            fmt.Printf(", count[%d]: %d", i, count[i]);
        }
        fmt.Printf("\r\n")

        if (loop == 10) {
            mymachine.SetRepeatedTimerAlarm("alarm0", mymachine.ALARM0, 1*1000*1000, alarm, 4, true)
            mymachine.SetOneshotTimerAlarm("alarm1", mymachine.ALARM1, 5*1000*1000, alarm, 5, true)
            mymachine.SetRepeatedTimerAlarm("alarm2", mymachine.ALARM2, 0.1*1000*1000, ledToggle, 6)
            mymachine.SetRepeatedTimerAlarm("alarm3", mymachine.ALARM3, 3*1000*1000, alarm, 7, true)
            fmt.Printf("setting changed 1\r\n");
        } else if (loop == 20) {
            mymachine.SetRepeatedTimerAlarm("alarm1", mymachine.ALARM1, 0, alarm, 8)
            fmt.Printf("setting changed 2\r\n");
        } else if (loop == 30) {
            // to check if ALARM3, which has lowest priority, fires in spite of too busy ALARM1
            mymachine.SetOneshotTimerAlarm("exit", mymachine.ALARM3, 5*1000*1000, exit)
            fmt.Printf("setting changed 3\r\n");
        }
    }
}

func ledToggle(name string, alarmId mymachine.AlarmId, opts ...interface{}) {
    led.Toggle()
    count[alarmId]++
}

func alarm(name string, alarmId mymachine.AlarmId, opts ...interface{}) {
    task := 0
    if len(opts) > 0 {
        task = opts[0].(int)
    }
    verbose := false
    if len(opts) > 1 {
        verbose = opts[1].(bool)
    }
    if verbose {
        fmt.Printf("%s (task%d)\r\n", name, task);
    }
    count[alarmId]++
}

func exit(name string, alarmId mymachine.AlarmId, opts ...interface{}) {
    fmt.Printf("%s\r\n", name);
    os.Exit(0)
}