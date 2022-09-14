package buttons

import (
    "math/bits"
)

type historyType uint64

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

func (history *historyType) unshift(flag bool) {
    *history = historyType((uint64(*history) << 1) | boolToUint64(flag))
}

func (history *historyType) countRisingEdge(single bool) (count uint8) {
    u64 := uint64(*history)
    const evenMask = 0x5555555555555555
    const oddMask  = 0xaaaaaaaaaaaaaaaa
    // check rising at even column
    u64e := u64 ^ evenMask
    u64e = (u64e >> 1) | u64e
    u64e = u64e & evenMask
    // check rising at odd column
    u64o := u64 ^ oddMask
    u64o = (u64o >> 1) | u64o
    u64o = u64o & oddMask
    // merge even and odd (ignore MSB)
    u64 = ^(u64o | u64e | (uint64(1) << 63))
    // shortcuts
    if u64 == uint64(0) {
        return uint8(0)
    } else if single {
        return uint8(1)
    }
    // now '1' indicates where rising edge is
    return uint8(bits.OnesCount64(u64))
}

func (history *historyType) recentStayReleasedCounts() uint8 {
    u64 := uint64(*history)
    return uint8(bits.TrailingZeros64(u64))
}

func (history *historyType) recentStayPushedCounts() uint8 {
    u64 := uint64(*history)
    // shortcut for very usual history case
    if u64 == uint64(0) {
        return uint8(0)
    }
    return uint8(bits.TrailingZeros64(^u64))
}
