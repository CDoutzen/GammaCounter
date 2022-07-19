package randstring

import (
	"math/rand"
	"strings"
)

var sets = "qwertyuiopasdfghjklzxcvbnm1234567890"

func RandString(lens int) string {
	bd := strings.Builder{}
	for i := 0; i < lens; i++ {
		bd.WriteByte(sets[rand.Intn(36)])
	}
	return bd.String()
}
