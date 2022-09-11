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

func (history *historyType) unshiftPos(flag bool) {
    *history = historyType((uint64(*history) << 1) | boolToUint64(flag))
}
