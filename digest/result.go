package digest

import (
	"fmt"
)

// Result is simply a byte slice, representing an SHA1 signature.
type Result []byte

func (r Result) String() string {
	return fmt.Sprintf("%x", []byte(r))
}
