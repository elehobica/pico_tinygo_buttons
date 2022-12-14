package main

import (
    "fmt"
    "machine"
    //"time"

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

var t uint64
var scanCnt uint32

func main() {
    println(); println()
    println("=========================")
    println("== pico_tinygo_buttons ==")
    println("=========================")

    led.Configure(machine.PinConfig{Mode: machine.PinOutput})
    led.Low()

    resetBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    setBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    centerBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    leftBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    rightBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    upBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
    downBtnPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

    btns := buttons.New("5WayTactile+2",
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

    err := mymachine.SetRepeatedTimerAlarm("alarm1", mymachine.ALARM1, 50*1000, buttonScan, btns)
    if err != nil {
        println(err)
        return
    }

    for loop := 0; true; loop++ {
        for event := btns.GetEvent(); event != nil; event = btns.GetEvent() {
            switch event.Type {
            case buttons.EVT_SINGLE:
                if event.RepeatCount > 0 {
                    fmt.Printf("%s: 1 (Repeated %d)\r\n", event.ButtonName, event.RepeatCount)
                } else {
                    fmt.Printf("%s: 1\r\n", event.ButtonName)
                }
            case buttons.EVT_MULTI:
                fmt.Printf("%s: %d\r\n", event.ButtonName, event.ClickCount)
                if event.ButtonName == "center" && event.ClickCount == 3 {
                    fmt.Printf("time %dus (scan: %d)\r\n", t, scanCnt)
                }
            case buttons.EVT_LONG:
                fmt.Printf("%s: Long\r\n", event.ButtonName)
            case buttons.EVT_LONG_LONG:
                fmt.Printf("%s: LongLong\r\n", event.ButtonName)
            }
        }
        led.Toggle()
        //time.Sleep(100 * time.Millisecond)
    }
}

func buttonScan(name string, alarmId mymachine.AlarmId, opts ...interface{}) {
    btns := opts[0].(*buttons.Buttons)
    t0 := mymachine.TimeElapsed()
    buttons.ScanPeriodic(btns)
    t = mymachine.TimeElapsed() - t0
    scanCnt++
}
