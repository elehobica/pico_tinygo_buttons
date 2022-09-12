package buttons

const ButtonEventChanSize = 16

type Buttons struct {
    name        string
    buttonSlice []*Button
    scanSkip    uint8
    scanCnt     uint32
    event       chan ButtonEvent
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

func ScanPeriodic(buttons *Buttons) {
    defer func() { buttons.scanCnt++ } ()
    if buttons.scanCnt < uint32(buttons.scanSkip) {
        return
    }
    for _, button := range buttons.buttonSlice {
        // what to get (default values)
        var repeatCnt, countRise uint8
        var detectLong, detectLongLong bool
        // alias
        cfg := button.config
        // === get raw status of pin ===
        rawSts := button.pin.Get() == cfg.activeHigh
        // === unshift history ===
        button.history.unshift(rawSts)
        trailingOnes := button.history.trailingOnes()
        trailingZeros :=button.history.trailingZeros()
        // === Detect Repeated (by non-filtered) ===
        if cfg.longDetectCnt == 0 && cfg.longLongDetectCnt == 0 {
            if buttons.scanCnt % uint32(cfg.repeatSkip + 1) == 0 {
                if cfg.repeatDetectCnt > 0 && trailingOnes >= cfg.repeatDetectCnt {
                    if button.rptCnt < 255 {
                        button.rptCnt++
                    }
                    repeatCnt = button.rptCnt
                } else {
                    button.rptCnt = 0
                }
            }
        }
        // === Detect Long (by non-filtered) ===
        if cfg.repeatDetectCnt == 0 {
            if trailingOnes > 0 {
                if trailingOnes == cfg.longDetectCnt {
                    detectLong = true
                } else if trailingOnes == cfg.longLongDetectCnt {
                    detectLongLong = true
                }
            }
        }
        // === unshift Filter ===
        if trailingOnes >= cfg.filterSize {
            button.filtered.unshift(true)
        } else if trailingZeros >= cfg.filterSize {
            button.filtered.unshift(false)
        } else {
            button.filtered.unshift(button.filtered.getPos(0))
        }
        // === Check Action finished (only if multiClicks) ===
        actFinished := button.filtered.trailingZeros() >= cfg.actFinishCnt
        // === Then, Count rising edge ===
        if repeatCnt > 0 { // if repeatCnt,countRise could be 0
            countRise = 1
        } else if actFinished {
            countRise = button.filtered.countRisingEdge(!cfg.multiClicks)
        }
        // Clear all once detected, initialize all as true to avoid repeated detection
        if detectLong || countRise > 0 {
            button.filtered = newHistory(true)
        }
        // === Send event ===
        eventType := EVT_NONE
        if countRise > 1 {
            eventType = EVT_MULTI
        } else if countRise > 0 {
            eventType = EVT_SINGLE
        } else if detectLong {
            eventType = EVT_LONG
        } else if detectLongLong {
            eventType = EVT_LONG_LONG
        }
        if eventType != EVT_NONE && len(buttons.event) < cap(buttons.event) {
            buttons.event <- ButtonEvent {
                ButtonName: button.name,
                Type: eventType,
                ClickCount: countRise,
                RepeatCount: repeatCnt,
            }
        }
    }
}
