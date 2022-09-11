package buttons

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

// trailingZeros
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.3:src/math/bits/bits.go;l=35-100
const deBruijn64 = 0x03f79d71b4ca8b09
var deBruijn64tab = [64]byte{
    0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
    62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
    63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
    54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

func trailingZeros(x uint64) uint8 {
    if x == 0 {
        return uint8(64)
    }
    return uint8(deBruijn64tab[(x&-x)*deBruijn64>>(64-6)])
}

func (history *historyType) trailingZeros() uint8 {
    u64 := uint64(*history)
    return trailingZeros(u64)
}

func (history *historyType) trailingOnes() uint8 {
    u64 := uint64(*history)
    // short cut for very usual history case
    if u64 == uint64(0) {
        return uint8(0)
    }
    return trailingZeros(^u64)
}
