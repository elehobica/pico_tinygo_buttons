// +build rp2040

package mymachine

import (
    "device/rp"
    "runtime/interrupt"
    "runtime/volatile"
    "unsafe"
    "fmt"
)

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

type AlarmId int
const (
    ALARM0 AlarmId = iota
    ALARM1
    ALARM2
    ALARM3
    __NUM_ALARMS__
)

type callbackType func(name string, alarmId AlarmId, opts ...interface{})

type Alarm struct {
    name      string
    intr      interrupt.Interrupt
    repeat    bool
    interval  uint32
    target    uint64
    callback  callbackType
    opts      []interface{}
}

const minInterval uint32 = 2 // us
var almAry [__NUM_ALARMS__]*Alarm

// TimeElapsed returns time elapsed since power up, in microseconds.
func (tmr *timerType) timeElapsed() (us uint64) {
    // As long as accessing order, "accessing the lower register, L, followed by the higher register, H", is kept,
    // the series of 64bit value is guranteed by the Pico Timer's hardware logic.
    lo := tmr.timeRawL.Get()
    hi := tmr.timeRawH.Get()
    return uint64(hi)<<32 | uint64(lo)
}

func TimeElapsed() (us uint64) {
    return timer.timeElapsed()
}

func setTimerIrq(alarmId AlarmId, flag bool) error {
    if alarmId >= __NUM_ALARMS__ {
        return fmt.Errorf("AlarmId over: %d", alarmId)
    }
    if alarmId == ALARM0 {
        return fmt.Errorf("ALARM%d is reserved by TinyGo machine rp2040", alarmId)
    }
    if (flag) {
        switch (alarmId) { // interrupt.New() only permits const value as IRQ
        case ALARM1:
            almAry[alarmId].intr = interrupt.New(rp.IRQ_TIMER_IRQ_1, timerHandleInterrupt1)
        case ALARM2:
            almAry[alarmId].intr = interrupt.New(rp.IRQ_TIMER_IRQ_2, timerHandleInterrupt2)
        case ALARM3:
            almAry[alarmId].intr = interrupt.New(rp.IRQ_TIMER_IRQ_3, timerHandleInterrupt3)
        }
        almAry[alarmId].intr.Enable()
        timer.intE.SetBits(1 << alarmId)
        irqSet(uint32(rp.IRQ_TIMER_IRQ_0 + int(alarmId)), true)
    } else {
        almAry[alarmId].intr.Disable()
        timer.intE.ClearBits(1 << alarmId)
        timer.armed.Set(1 << alarmId) // Disarm timer
        irqSet(uint32(rp.IRQ_TIMER_IRQ_0 + int(alarmId)), false)
    }
    return nil
}

func setTimerAlarm(name string, alarmId AlarmId, us uint32, repeat bool, callback callbackType, opts ...interface{}) (err error) {
    if alarmId >= __NUM_ALARMS__ {
        return fmt.Errorf("AlarmId over")
    }

    if almAry[alarmId] == nil {
        almAry[alarmId] = &Alarm{}
    }
    if us < minInterval {
        us = minInterval
    }
    almAry[alarmId].name = name
    almAry[alarmId].repeat = repeat
    almAry[alarmId].interval = us
    almAry[alarmId].target = timer.timeElapsed() + uint64(almAry[alarmId].interval)
    almAry[alarmId].callback = callback
    almAry[alarmId].opts = opts

    if almAry[alarmId].callback != nil {
        now := timer.timeElapsed()
        // Care the case if the time has already passed the target
        if almAry[alarmId].target <= now {
            almAry[alarmId].target = now + uint64(minInterval)
        }
        timer.alarm[alarmId].Set(uint32(almAry[alarmId].target))
        err = setTimerIrq(alarmId, true)
    } else {
        almAry[alarmId].repeat = false
        err = setTimerIrq(alarmId, false)
    }

    return err
}

func SetRepeatedTimerAlarm(name string, alarmId AlarmId, us uint32, callback callbackType, opts ...interface{}) error {
    return setTimerAlarm(name, alarmId, us, true, callback, opts...)
}

func SetOneshotTimerAlarm(name string, alarmId AlarmId, us uint32, callback callbackType, opts ...interface{}) error {
    return setTimerAlarm(name, alarmId, us, false, callback, opts...)
}

func timerHandleInterrupt1(intr interrupt.Interrupt) {
    timerHandleInterrupt(ALARM1, intr)
}

func timerHandleInterrupt2(intr interrupt.Interrupt) {
    timerHandleInterrupt(ALARM2, intr)
}

func timerHandleInterrupt3(intr interrupt.Interrupt) {
    timerHandleInterrupt(ALARM3, intr)
}

func timerHandleInterrupt(alarmId AlarmId, intr interrupt.Interrupt) {
    ints := timer.intS.Get()
    // Check if the IRQ fired
    if (ints & (1 << alarmId)) == 0 {
        return
    }
    // Clear Interrupt
    timer.intR.Set(1 << alarmId)
    if almAry[alarmId] == nil {
        return
    }
    if almAry[alarmId].callback == nil {
        return
    }
    // Do callback
    almAry[alarmId].callback(almAry[alarmId].name, alarmId, almAry[alarmId].opts...)
    // Prepare for repeated alarm
    if almAry[alarmId].repeat {
        almAry[alarmId].target += uint64(almAry[alarmId].interval)
        now := timer.timeElapsed()
        // Care the case if the time has already passed the target while callback
        if almAry[alarmId].target < now + uint64(minInterval) {
            almAry[alarmId].target = now + uint64(minInterval)
        }
        timer.alarm[alarmId].Set(uint32(almAry[alarmId].target))
        // it fires even if other callback(s) takes time
    } else {
        almAry[alarmId].callback = nil
        setTimerIrq(alarmId, false)
        intr.Disable()
    }
}
