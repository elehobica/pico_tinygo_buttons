package main

import (
    "fmt"
    "machine"
    "time"

    "github.com/elehobica/pico_tinygo_buttons/mymachine"
    "github.com/elehobica/pico_tinygo_buttons/buttons"
)

var (
    serial     = machine.Serial
    ledPin       machine.Pin
    resetBtnPin  machine.Pin
    setBtnPin    machine.Pin
    centerBtnPin machine.Pin
    rightBtnPin  machine.Pin
    leftBtnPin   machine.Pin
    downBtnPin   machine.Pin
    upBtnPin     machine.Pin
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
    println("=========================")
    println("== pico_tinygo_buttons ==")
    println("=========================")

    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    led.Low()

    btns := buttons.New("5WayTactile+2")
    btns.AddButton(
        []*buttons.Button {
            buttons.NewButton("reset",  &resetBtnPin,  buttons.DefaultButtonSingleConfig),
            buttons.NewButton("set",    &setBtnPin,    buttons.DefaultButtonSingleConfig),
            buttons.NewButton("center", &centerBtnPin, buttons.DefaultButtonMultiConfig),
            buttons.NewButton("left",   &leftBtnPin,   buttons.DefaultButtonSingleRepeatConfig),
            buttons.NewButton("right",  &rightBtnPin,  buttons.DefaultButtonSingleRepeatConfig),
            buttons.NewButton("up",     &upBtnPin,     buttons.DefaultButtonSingleRepeatConfig),
            buttons.NewButton("down",   &downBtnPin,   buttons.DefaultButtonSingleRepeatConfig),
        }...
    );

    mymachine.SetRepeatedTimerAlarm("alarm0", mymachine.ALARM0, 0.05*1000*1000, buttonScan, btns)

    for loop := 0; true; loop++ {
        if event := btns.GetEvent(); event != nil {
            if event.Type == buttons.EVT_SINGLE {
                fmt.Printf("%s: 1\r\n", event.Button.GetName())
            } else if event.Type == buttons.EVT_SINGLE_REPEATED {
                fmt.Printf("%s: 1 (Repeated)\r\n", event.Button.GetName())
            } else if event.Type == buttons.EVT_MULTI {
                fmt.Printf("%s: %d\r\n", event.Button.GetName(), event.Count)
            } else if event.Type == buttons.EVT_LONG {
                fmt.Printf("%s: Long\r\n", event.Button.GetName())
            } else if event.Type == buttons.EVT_LONG_LONG {
                fmt.Printf("%s: LongLong\r\n", event.Button.GetName())
            }
        }
        led.Toggle()
        time.Sleep(100 * time.Millisecond)
    }
}

func buttonScan(name string, alarmId mymachine.AlarmId, opts ...interface{}) {
    btns := opts[0].(*buttons.Buttons)
    //t0 := mymachine.TimeElapsed()
    buttons.ScanPeriodic(btns)
    //t1 := mymachine.TimeElapsed()
    //fmt.Printf("time %d\r\n", t1 - t0)
}
