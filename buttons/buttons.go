package buttons

import (
    "machine"
)

type ButtonConfig struct {
    activeHigh bool         // Set false if button is connected between GND and pin with pull-up
    multiClicks bool        // Detect multiple clicks if true, detect single click if false
    historySize uint8       // Size of button status history (max 64)
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

type historyType uint64

type Button struct {
    name     string
    pin      *machine.Pin
    config   *ButtonConfig
    history  historyType
    filtered historyType
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
    ButtonName  string
    Type        ButtonEventType
    ClickCount  uint8
    RepeatCount uint8
}

const ButtonEventChanSize = 16

type Buttons struct {
    name        string
    buttonSlice []*Button
    scanSkip    uint8
    scanCnt     uint32
    event       chan ButtonEvent
}

//=============
// historyType
//=============
func newHistory(flag bool) historyType {
    if flag {
        return historyType(^uint64(0))
    }
    return historyType(uint64(0))
}

func boolToUint64(flag bool) (ans uint64) {
    if flag {
        ans = uint64(1)
    }
    return ans
}

func (history *historyType) getPos(i int) bool {
    mask := uint64(1) << i
    return (uint64(*history) & mask) != uint64(0)
}

func (history *historyType) setPos(i int, flag bool) {
    mask := uint64(1) << i
    *history = historyType((uint64(*history) & ^mask) | (boolToUint64(flag) << i))
}

func (history *historyType) unshiftPos(flag bool) {
    *history = historyType((uint64(*history) << 1) | boolToUint64(flag))
}

//==============
// ButtonCOnfig
//==============
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
    } else if config.historySize > 64 {
        config.historySize = 64
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

//=========
// Buttons
//=========
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

//========
// Button
//========
func NewButton(name string, pin *machine.Pin, config *ButtonConfig) *Button {
    button := Button {
        name: name,
        pin: pin,
        config: config,
        history: newHistory(false),
        filtered: newHistory(false),
    }
    button.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
    return &button
}

//===============
// Scan Periodic
//===============
func ScanPeriodic(buttons *Buttons) {
    defer func() { buttons.scanCnt++ } ()
    if buttons.scanCnt < uint32(buttons.scanSkip) {
        return
    }
    for _, button := range buttons.buttonSlice {
        pin := button.pin
        rawSts := pin.Get() == button.config.activeHigh
        // === unshift with keeping slice size ===
        button.history.unshiftPos(rawSts)
        button.filtered.unshiftPos(false)
        // === Detect Repeated (by non-filtered) ===
        repeatCnt := func(history *historyType, config *ButtonConfig) (repeat uint8) {
            if config.longDetectCnt == 0 && config.longLongDetectCnt == 0 {
                var count uint8 = 0
                for i := 0; i < int(config.historySize); i++ {
                    if history.getPos(i) {
                        count++
                    } else {
                        break;
                    }
                }
                if buttons.scanCnt % uint32(config.repeatSkip + 1) == 0 {
                    if config.repeatDetectCnt > 0 && count >= config.repeatDetectCnt {
                        repeat = (count - config.repeatDetectCnt + config.repeatSkip + 1) / (config.repeatSkip + 1)
                    }
                }
            }
            return repeat
        } (&button.history, button.config)
        // === Detect Long (by non-filtered) ===
        detectLong, detectLongLong := func(history, filtered *historyType, config *ButtonConfig) (flagLong, flagLongLong bool) {
            if config.repeatDetectCnt == 0 {
                var count uint8 = 0
                for i := 0; i < int(config.historySize); i++ {
                    if history.getPos(i) {
                        count++
                    } else {
                        break;
                    }
                }
                if count > 0 {
                    if count == config.longDetectCnt {
                        flagLong = true
                    } else if count == config.longLongDetectCnt {
                        flagLongLong = true
                    }
                }
                if flagLong {
                    // Clear all once detected, initialize all as true to avoid repeated detection
                    *filtered = newHistory(true)
                }
            }
            return flagLong, flagLongLong
        } (&button.history, &button.filtered, button.config)
        // === Filter ===
        {
            isFilteredTrue := true
            isFilteredFalse := true
            for i := 0; i < int(button.config.filterSize); i++ {
                histSts := button.history.getPos(i)
                isFilteredTrue = isFilteredTrue && histSts
                isFilteredFalse = isFilteredFalse && !histSts
            }
            if isFilteredTrue {
                button.filtered.setPos(0, true)
            } else if isFilteredFalse {
                button.filtered.setPos(0, false)
            } else {
                button.filtered.setPos(0, button.filtered.getPos(1))
            }
        }
        // === Check Action finished (only if multiClicks) ===
        actFinished := func(filtered *historyType, config *ButtonConfig) (flag bool) {
            flag = true
            for i := 0; i < int(config.actFinishCnt); i++ {
                flag = flag && !filtered.getPos(i)
            }
            return flag
        } (&button.filtered, button.config)
        // === Count rising edge ===
        countRise := func(filtered *historyType, config *ButtonConfig) (count uint8) {
            if actFinished {
                for i := 0; i < int(config.historySize - 1); i++ {
                    if filtered.getPos(i) && !filtered.getPos(i+1) {
                        count++
                        if !config.multiClicks {
                            break
                        }
                    }
                }
                if count > 0 {
                    // Clear all once detected, initialize all as true to avoid repeated detection
                    *filtered = newHistory(true)
                }
            }
            return count
        } (&button.filtered, button.config)
        // === Send event ===
        {
            eventType := EVT_NONE
            if repeatCnt > 0 {
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
                ClickCount: countRise,
                RepeatCount: repeatCnt,
            }
            if eventType != EVT_NONE && len(buttons.event) < cap(buttons.event) {
                buttons.event <- event
            }
        }
    }
}
