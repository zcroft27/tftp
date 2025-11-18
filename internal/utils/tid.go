package utils

import (
	"math/rand/v2"
)

func GenerateTID() int {
	TID := 49152 + rand.IntN(65536-49152) // [49152, 65535] is suggested in RFC 6335 as ephemeral ports for dynamic assignment.
	return TID
}
