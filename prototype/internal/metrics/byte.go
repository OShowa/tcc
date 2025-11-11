package metrics

import (
	"fmt"
	"math"
)

type Byte int64

func KB(kb int64) Byte {
	return Byte(kb * 1e3)
}

func MB(mb int64) Byte {
	return Byte(mb * 1e6)
}

func GB(gb int64) Byte {
	return Byte(gb * 1e9)
}

func TB(tb int64) Byte {
	return Byte(tb * 1e12)
}

func (bs Byte) String() string {
	magnitude := int64(math.Log10(float64(bs)))
	if magnitude < 3 {
		return fmt.Sprintf("%dB", bs)
	}
	if magnitude < 6 {
		return fmt.Sprintf("%.2fkB", float64(bs)/1e3)
	}
	if magnitude < 9 {
		return fmt.Sprintf("%.2fMB", float64(bs)/1e6)
	}
	if magnitude < 12 {
		return fmt.Sprintf("%.2fGB", float64(bs)/1e9)
	}
	return fmt.Sprintf("%.2fTB", float64(bs)/1e12)
}
