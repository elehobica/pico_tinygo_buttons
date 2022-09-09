// +build rp2040

package mymachine

import (
    "device/rp"
    "runtime/interrupt"
    "runtime/volatile"
    "unsafe"
    "sync"
    "fmt"
)

type AlarmId int
const (
    ALARM0 AlarmId = iota
    ALARM1
    ALARM2
    ALARM3
    __NUM_ALARMS__
)

type Alarm struct {
    repeat    bool
    interval  uint32
    target    uint32
    isPending bool
    callback  func()
}

var alarmAry [__NUM_ALARMS__]*Alarm

type timerType struct {
    timeHW   volatile.Register32
    timeLW   volatile.Register32
    timeHR   volatile.Register32
    timeLR   volatile.Register32
    alarm    [__NUM_ALARMS__]volatile.Register32
    armed    volatile.Register32
    timeRawH volatile.Register32
    timeRawL volatile.Register32
    dbgPause volatile.Register32
    pause    volatile.Register32
    intR     volatile.Register32
    intE     volatile.Register32
    intF     volatile.Register32
    intS     volatile.Register32
}

var timer = (*timerType)(unsafe.Pointer(rp.TIMER))
var mu = (*sync.Mutex)(&sync.Mutex{})

// TimeElapsed returns time elapsed since power up, in microseconds.
func (tmr *timerType) timeElapsed() (us uint64) {
    // Need to make sure that the upper 32 bits of the timer
    // don't change, so read that first
    hi := tmr.timeRawH.Get()
    var lo, nextHi uint32
    for {
        // Read the lower 32 bits
        lo = tmr.timeRawL.Get()
        // Now read the upper 32 bits again and
        // check that it hasn't incremented. If it has, loop around
        // and read the lower 32 bits again to get an accurate value
        nextHi = tmr.timeRawH.Get()
        if hi == nextHi {
            break
        }
        hi = nextHi
    }
    return uint64(hi)<<32 | uint64(lo)
}

func setTimerIrq(alarmId AlarmId, flag bool) error{
    if alarmId >= __NUM_ALARMS__ {
        return fmt.Errorf("AlarmId over: %d", alarmId)
    }
    mu.Lock()
    defer mu.Unlock()
    inte := timer.intE.Get()
    if (flag) {
        timer.intE.Set(inte | (1 << alarmId))
        switch (alarmId) { // interrupt.New() only permits const value as IRQ
        case ALARM0:
            interrupt.New(rp.IRQ_TIMER_IRQ_0, timerHandleInterrupt).Enable()
        case ALARM1:
            interrupt.New(rp.IRQ_TIMER_IRQ_1, timerHandleInterrupt).Enable()
        case ALARM2:
            interrupt.New(rp.IRQ_TIMER_IRQ_2, timerHandleInterrupt).Enable()
        case ALARM3:
            interrupt.New(rp.IRQ_TIMER_IRQ_3, timerHandleInterrupt).Enable()
        }
        irqSet(uint32(rp.IRQ_TIMER_IRQ_0 + int(alarmId)), true)
    } else {
        timer.intE.Set(inte & ((1 << alarmId) ^ 0xffffffff))
        switch (alarmId) { // interrupt.New() only permits const value as IRQ
        case ALARM0:
            interrupt.New(rp.IRQ_TIMER_IRQ_0, timerHandleInterrupt).Disable()
        case ALARM1:
            interrupt.New(rp.IRQ_TIMER_IRQ_1, timerHandleInterrupt).Disable()
        case ALARM2:
            interrupt.New(rp.IRQ_TIMER_IRQ_2, timerHandleInterrupt).Disable()
        case ALARM3:
            interrupt.New(rp.IRQ_TIMER_IRQ_3, timerHandleInterrupt).Disable()
        }
        irqSet(uint32(rp.IRQ_TIMER_IRQ_0 + int(alarmId)), false)
    }
    return nil
}

func setTimerAlarm(alarmId AlarmId, us uint32, repeat bool, f func()) error {
    if alarmId >= __NUM_ALARMS__ {
        return fmt.Errorf("AlarmId over")
    }

    if alarmAry[alarmId] == nil {
        alarmAry[alarmId] = &Alarm{}
    }
    alarmAry[alarmId].repeat = repeat
    alarmAry[alarmId].interval = us
    alarmAry[alarmId].target = timer.timeRawL.Get() + us
    alarmAry[alarmId].isPending = false
    alarmAry[alarmId].callback = f

    if alarmAry[alarmId].callback != nil {
        setTimerIrq(alarmId, true)
        timer.alarm[alarmId].Set(alarmAry[alarmId].target)
    } else {
        setTimerIrq(alarmId, false)
    }

    return nil
}

func SetRepeatedTimerAlarm(alarmId AlarmId, us uint32, f func()) error {
    return setTimerAlarm(alarmId, us, true, f)
}

func SetOneshotTimerAlarm(alarmId AlarmId, us uint32, f func()) error {
    return setTimerAlarm(alarmId, us, false, f)
}

func timerHandleInterrupt(intr interrupt.Interrupt) {
    ints := timer.intS.Get()
    for i := 0; i < int(__NUM_ALARMS__); i++ {
        if (ints & (1 << i)) != 0 {
            mu.Lock()
            timer.intR.Set(1 << i) // Clear Interrupt
            mu.Unlock()
            if alarmAry[i] != nil {
                alarmAry[i].isPending = true
                if alarmAry[i].callback != nil {
                    alarmAry[i].callback()
                }
                if alarmAry[i].repeat {
                    alarmAry[i].target += alarmAry[i].interval
                    timer.alarm[i].Set(alarmAry[i].target)
                    // TO DO
                    // Need to care the case if the time passed the target while callback
                } else {
                    alarmAry[i].callback = nil
                    setTimerIrq(AlarmId(i), false)
                }
                alarmAry[i].isPending = false
            }
            break // process only one IRQ (priority: ALARM0 > ALARM1 > ALARM2 > ALARM3)
        }
    }
}
