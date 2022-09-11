package buttons

import (
    "machine"
)

type ButtonConfig struct {
    activeHigh bool         // Set false if button is connected between GND and pin with pull-up
    multiClicks bool        // Detect multiple clicks if true, detect single click if false
    historySize uint8       // Size of button status history
    filterSize uint8        // filter size to process raw status
    actFinishCnt uint8      // Button action detection starts when status keeps false at latest continuous actFinishCnt times (only if multiClicks)
    repeatDetectCnt uint8   // continuous counts to detect Repeat click when continuous push (only if !multiClicks. ignored if 0. if defined Long/LongLong detect disabled)
    repeatSkip uint8        // skip count for Repeat click detection (every scan if 0)
    longDetectCnt uint8     // continuous counts to detect Long Push (ignored if 0)
    longLongDetectCnt uint8 // continuous counts to detect LongLong Push (ignored if 0)
}

var DefaultButtonSingleConfig = &ButtonConfig {
    activeHigh: false,
    multiClicks: false,
    historySize: 40,
    filterSize: 1,
    actFinishCnt: 0,
    repeatDetectCnt: 0,
    repeatSkip: 0,
    longDetectCnt: 0,
    longLongDetectCnt: 0,
}

var DefaultButtonSingleRepeatConfig = &ButtonConfig {
    activeHigh: false,
    multiClicks: false,
    historySize: 40,
    filterSize: 1,
    actFinishCnt: 0,
    repeatDetectCnt: 10,
    repeatSkip: 2,
    longDetectCnt: 0,
    longLongDetectCnt: 0,
}

var DefaultButtonMultiConfig = &ButtonConfig {
    activeHigh: false,
    multiClicks: true,
    historySize: 40,
    filterSize: 1,
    actFinishCnt: 5,
    repeatDetectCnt: 0,
    repeatSkip: 2,
    longDetectCnt: 15,
    longLongDetectCnt: 39,
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
    ButtonName string
    Type       ButtonEventType
    Count      uint8
}

const ButtonEventChanSize int = 16

type Buttons struct {
    name        string
    buttonSlice []*Button
    scanSkip    uint8
    scanCnt     uint32
    event       chan ButtonEvent
}

func NewButtonConfig(
        activeHigh, multiClicks bool,
        historySize, filterSize, actFinishCnt, repeatDetectCnt, repeatSkip, longDetectCnt, longLongDetectCnt uint8,
    ) *ButtonConfig {
    config := &ButtonConfig {
        activeHigh: activeHigh,
        multiClicks: multiClicks,
        historySize: historySize,
        filterSize: filterSize,
        actFinishCnt: actFinishCnt,
        repeatDetectCnt: repeatDetectCnt,
        repeatSkip: repeatSkip,
        longDetectCnt: longDetectCnt,
        longLongDetectCnt: longLongDetectCnt,
    }
    config.reflectConstraints()
    return config
}

func (config *ButtonConfig) reflectConstraints() {
    // revise illegal settings
    if config.historySize < 10 {
        config.historySize = 10
    }
    if config.filterSize < 1 {
        config.filterSize = 1
    }
    if !config.multiClicks {
        config.actFinishCnt = 0
    } else if config.actFinishCnt > config.historySize {
        config.actFinishCnt = config.historySize
    }
    if config.longDetectCnt > config.historySize - 1{
        config.longDetectCnt = config.historySize - 1
    }
    if config.longLongDetectCnt > config.historySize - 1{
        config.longLongDetectCnt = config.historySize - 1
    }
}

func New(name string, button ...*Button) *Buttons {
    return &Buttons {
        name: name,
        buttonSlice: append([]*Button{}, button...),
        event: make(chan ButtonEvent, ButtonEventChanSize),
    }
}

func (buttons *Buttons) SetScanSkip(scanSkip uint8) {
    buttons.scanSkip = scanSkip
}

func (buttons *Buttons) GetName() string {
    return buttons.name
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

func NewButton(name string, pin *machine.Pin, config *ButtonConfig) *Button {
    button := Button {
        name: name,
        pin: pin,
        config: config,
        history: make([]bool, config.historySize, config.historySize),
        filtered: make([]bool, config.historySize, config.historySize),
    }
    button.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
    return &button
}

func ScanPeriodic(buttons *Buttons) {
    clearBoolSlice := func(slice []bool, value bool) {
        for i := range slice  {
            slice[i] = value
        }
    }

    defer func() { buttons.scanCnt++ } ()
    if buttons.scanCnt < uint32(buttons.scanSkip) {
        return
    }
    for _, button := range buttons.buttonSlice {
        pin := button.pin
        rawSts := pin.Get() == button.config.activeHigh
        // === unshift with keeping slice size ===
        {
            button.history = append([]bool{rawSts,}, button.history[:len(button.history)-1]...)
            button.filtered = append([]bool{false,}, button.filtered[:len(button.filtered)-1]...)
        }
        // === Detect Repeated (by non-filtered) ===
        detectRepeat := func() bool {
            detectRepeat := false
            if button.config.longDetectCnt == 0 && button.config.longLongDetectCnt == 0 {
                var count uint8 = 0
                for _, histSts := range button.history {
                    if histSts {
                        count++
                    } else {
                        break
                    }
                }
                if button.config.repeatDetectCnt > 0 && count >= button.config.repeatDetectCnt {
                    if buttons.scanCnt % uint32(button.config.repeatSkip + 1) == 0 {
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
            if button.config.repeatDetectCnt == 0 {
                var count uint8 = 0
                for _, histSts := range button.history {
                    if histSts {
                        count++
                    } else {
                        break
                    }
                }
                if count > 0 {
                    if count == button.config.longDetectCnt {
                        detectLong = true
                    } else if count == button.config.longLongDetectCnt {
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
            for _, histSts := range button.history[:button.config.filterSize] {
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
        // === Check Action finished (only if multiClicks) ===
        actFinished := func() bool {
            actFinished := true
            for _, filtSts := range button.filtered[:button.config.actFinishCnt] {
                actFinished = actFinished && !filtSts
            }
            return actFinished
        } ()
        // === Count rising edge ===
        countRise := func() int {
            countRise := 0
            if actFinished {
                for i := 0; i < int(button.config.historySize - 1); i++ {
                    if button.filtered[i] && !button.filtered[i+1] {
                        countRise++
                        if !button.config.multiClicks {
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
                ButtonName: button.name,
                Type: eventType,
                Count: uint8(countRise),
            }
            if eventType != EVT_NONE && len(buttons.event) < cap(buttons.event) {
                buttons.event <- event
            }
        }
    }
}
