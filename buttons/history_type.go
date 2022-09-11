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

func (history *historyType) countRisingEdge(size uint8, single bool) (count uint8) {
    for i := 0; i < int(size) - 1; i++ {
        if history.getPos(i) && !history.getPos(i+1) {
            count++
            if single {
                break
            }
        }
    }
    return count
}

func (history *historyType) trailingZeros() uint8 {
    u64 := uint64(*history)
    return uint8(bits.TrailingZeros64(u64))
}

func (history *historyType) trailingOnes() uint8 {
    u64 := uint64(*history)
    // short cut for very usual history case
    if u64 == uint64(0) {
        return uint8(0)
    }
    return uint8(bits.TrailingZeros64(^u64))
}
