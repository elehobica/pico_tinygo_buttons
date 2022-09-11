package buttons

import (
    "machine"
)

type ButtonConfig struct {
    ActiveHigh bool         // Set false if button is connected between GND and pin with pull-up
    MultiClicks bool        // Detect multiple clicks if true, detect single click if false
    HistorySize uint8       // Size of button status history
    FilterSize uint8        // filter size to process raw status
    ActFinishCnt uint8      // Button action detection starts when status keeps false at latest continuous ActFinishCnt times (only if MultiClicks)
    RepeatDetectCnt uint8   // continuous counts to detect Repeat click when continuous push (only if !MultiClicks. ignored if 0. if defined Long/LongLong detect disabled)
    RepeatSkip uint8        // skip count for Repeat click detection (every scan if 0)
    LongDetectCnt uint8     // continuous counts to detect Long Push (ignored if 0)
    LongLongDetectCnt uint8 // continuous counts to detect LongLong Push (ignored if 0)
}

var DefaultButtonSingleConfig = &ButtonConfig {
    ActiveHigh: false,
    MultiClicks: false,
    HistorySize: 40,
    FilterSize: 1,
    ActFinishCnt: 5,
    RepeatDetectCnt: 0,
    RepeatSkip: 0,
    LongDetectCnt: 0,
    LongLongDetectCnt: 0,
}

var DefaultButtonSingleRepeatConfig = &ButtonConfig {
    ActiveHigh: false,
    MultiClicks: false,
    HistorySize: 40,
    FilterSize: 1,
    ActFinishCnt: 5,
    RepeatDetectCnt: 10,
    RepeatSkip: 2,
    LongDetectCnt: 0,
    LongLongDetectCnt: 0,
}

var DefaultButtonMultiConfig = &ButtonConfig {
    ActiveHigh: false,
    MultiClicks: true,
    HistorySize: 40,
    FilterSize: 1,
    ActFinishCnt: 5,
    RepeatDetectCnt: 0,
    RepeatSkip: 2,
    LongDetectCnt: 15,
    LongLongDetectCnt: 39,
}

type Button struct {
    name     string
    pin      *machine.Pin
    config   *ButtonConfig
    history  []bool
    filtered []bool
}

type ButtonEventType int
const (
    EVT_NONE ButtonEventType = iota
    EVT_SINGLE
    EVT_SINGLE_REPEATED
    EVT_MULTI
    EVT_LONG
    EVT_LONG_LONG
)

type ButtonEvent struct {
    Button *Button
    Type   ButtonEventType
    Count  uint8
}

const ButtunEventSize int = 16

type Buttons struct {
    Name        string
    ButtonSlice []*Button
    ScanSkip    uint8
    scanCnt     uint32
    event       chan ButtonEvent
}

func (config *ButtonConfig) Configure() {
    // revise illegal settings
    if config.HistorySize < 10 {
        config.HistorySize = 10
    }
    if config.FilterSize < 1 {
        config.FilterSize = 1
    }
    if !config.MultiClicks {
        config.ActFinishCnt = 0
    } else if config.ActFinishCnt > config.HistorySize {
        config.ActFinishCnt = config.HistorySize
    }
    if config.LongDetectCnt > config.HistorySize - 1{
        config.LongDetectCnt = config.HistorySize - 1
    }
    if config.LongLongDetectCnt > config.HistorySize - 1{
        config.LongLongDetectCnt = config.HistorySize - 1
    }
}

func New(name string) *Buttons {
    return &Buttons {
        Name: name,
        event: make(chan ButtonEvent, ButtunEventSize),
    }
}

func (buttons *Buttons) AddButton(button ...*Button) {
    for _, v := range button {
        buttons.ButtonSlice = append(buttons.ButtonSlice, v)
    }
}

func (buttons *Buttons) GetEvent() *ButtonEvent {
    if len(buttons.event) == 0 {
        return nil
    }
    event, more := <- buttons.event
    if !more {
        return nil
    }
    return &event
}

func (button *Button) GetName() string {
    return button.name
}

func NewButton(name string, pin *machine.Pin, config *ButtonConfig) *Button {
    button := Button {
        name: name,
        pin: pin,
        config: config,
        history: make([]bool, config.HistorySize, config.HistorySize),
        filtered: make([]bool, config.HistorySize, config.HistorySize),
    }
    button.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
    button.config.Configure()
    return &button
}

func ScanPeriodic(buttons *Buttons) {
    clearBoolSlice := func(slice []bool, value bool) {
        for i := range slice  {
            slice[i] = value
        }
    }

    defer func() { buttons.scanCnt++ } ()
    if buttons.scanCnt < uint32(buttons.ScanSkip) {
        return
    }
    for _, button := range buttons.ButtonSlice {
        pin := button.pin
        rawSts := pin.Get() == button.config.ActiveHigh
        // === unshift with keeping slice size ===
        {
            button.history = append([]bool{rawSts,}, button.history[:len(button.history)-1]...)
            button.filtered = append([]bool{false,}, button.filtered[:len(button.filtered)-1]...)
        }
        // === Detect Repeated (by non-filtered) ===
        detectRepeat := func() bool {
            detectRepeat := false
            if button.config.LongDetectCnt == 0 && button.config.LongLongDetectCnt == 0 {
                var count uint8 = 0
                for _, histSts := range button.history {
                    if histSts {
                        count++
                    } else {
                        break
                    }
                }
                if button.config.RepeatDetectCnt > 0 && count >= button.config.RepeatDetectCnt {
                    if buttons.scanCnt % uint32(button.config.RepeatSkip + 1) == 0 {
                        detectRepeat = true
                    }
                }
            }
            return detectRepeat
        } ()
        // === Detect Long (by non-filtered) ===
        detectLong, detectLongLong := func() (bool, bool) {
            detectLong := false
            detectLongLong := false
            if button.config.RepeatDetectCnt == 0 {
                var count uint8 = 0
                for _, histSts := range button.history {
                    if histSts {
                        count++
                    } else {
                        break
                    }
                }
                if count > 0 {
                    if count == button.config.LongDetectCnt {
                        detectLong = true
                    } else if count == button.config.LongLongDetectCnt {
                        detectLongLong = true
                    }
                }
                if detectLong {
                    // Clear all once detected, initialize all as true to avoid repeated detection
                    clearBoolSlice(button.filtered, true)
                }
            }
            return detectLong, detectLongLong
        } ()
        // === Filter ===
        {
            isFilteredTrue := true
            isFilteredFalse := true
            for _, histSts := range button.history[:button.config.FilterSize] {
                isFilteredTrue = isFilteredTrue && histSts
                isFilteredFalse = isFilteredFalse && !histSts
            }
            if isFilteredTrue {
                button.filtered[0] = true
            } else if isFilteredFalse {
                button.filtered[0] = false
            } else {
                button.filtered[0] = button.filtered[1]
            }
        }
        // === Check Action finished (only if MultiClicks) ===
        actFinished := func() bool {
            actFinished := true
            for _, filtSts := range button.filtered[:button.config.ActFinishCnt] {
                actFinished = actFinished && !filtSts
            }
            return actFinished
        } ()
        // === Count rising edge ===
        countRise := func() int {
            countRise := 0
            if actFinished {
                for i := 0; i < int(button.config.HistorySize - 1); i++ {
                    if button.filtered[i] && !button.filtered[i+1] {
                        countRise++
                        if !button.config.MultiClicks {
                            break
                        }
                    }
                }
                if countRise > 0 {
                    // Clear all once detected, initialize all as true to avoid repeated detection
                    clearBoolSlice(button.filtered, true)
                }
            }
            return countRise
        } ()
        // === Send event ===
        {
            eventType := EVT_NONE
            if detectRepeat {
                eventType = EVT_SINGLE_REPEATED
            } else if countRise > 1 {
                eventType = EVT_MULTI
            } else if countRise > 0 {
                eventType = EVT_SINGLE
            } else if detectLong {
                eventType = EVT_LONG
            } else if detectLongLong {
                eventType = EVT_LONG_LONG
            }
            event := ButtonEvent {
                Button: button,
                Type: eventType,
                Count: uint8(countRise),
            }
            if eventType != EVT_NONE && len(buttons.event) < cap(buttons.event) {
                buttons.event <- event
            }
        }
    }
}
