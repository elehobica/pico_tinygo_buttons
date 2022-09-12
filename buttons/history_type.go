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
    // check rising at even colum
    u64l := u64 ^ evenMask
    u64l = (u64l >> 1) | u64l
    u64l = u64l & evenMask
    // check rising at odd colum
    u64h := u64 ^ oddMask 
    u64h = (u64h >> 1) | u64h
    u64h = u64h & oddMask
    // merge even and odd (ignore MSB)
    u64 = ^(u64h | u64l | (uint64(1) << 63))
    // shortcuts
    if u64 == uint64(0) {
        return uint8(0)
    } else if single {
        return uint8(1)
    }
    // countOnes where indicates rising edge as '1'
    return uint8(bits.OnesCount64(u64))
}

func (history *historyType) trailingZeros() uint8 {
    u64 := uint64(*history)
    return uint8(bits.TrailingZeros64(u64))
}

func (history *historyType) trailingOnes() uint8 {
    u64 := uint64(*history)
    // shortcut for very usual history case
    if u64 == uint64(0) {
        return uint8(0)
    }
    return uint8(bits.TrailingZeros64(^u64))
}
