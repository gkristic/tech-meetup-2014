package main

import (
	"fmt"
)

type size int64

// Units for size.
const (
	B = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
	PiB
)

func (n size) String() string {
	suffix := "B"
	factor := int64(1)

	switch {
	case n > PiB:
		suffix = "PiB"
		factor = PiB
	case n > TiB:
		suffix = "TiB"
		factor = TiB
	case n > GiB:
		suffix = "GiB"
		factor = GiB
	case n > MiB:
		suffix = "MiB"
		factor = MiB
	case n > KiB:
		suffix = "KiB"
		factor = KiB
	default:
		// Bytes; avoiding decimals
		return fmt.Sprintf("%dB", int64(n))
	}

	return fmt.Sprintf("%.2f%s", float64(n)/float64(factor), suffix)
}
