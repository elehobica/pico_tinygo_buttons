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
        pin := button.pin
        rawSts := pin.Get() == button.config.activeHigh
        // === unshift with keeping size ===
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
