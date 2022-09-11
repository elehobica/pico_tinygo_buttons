package buttons

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
