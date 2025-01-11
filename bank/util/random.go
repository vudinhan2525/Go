package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "qwertyuiopasdfghjklzxcvbnm"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomStr(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < k; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomStr(6)
}
func RandomMoney() int64 {
	return RandomInt(0, 200000)
}
func RandomCurrency() string {
	currencies := []string{"VND", "USD", "EUR"}
	n := len(currencies)

	return currencies[rand.Intn(n)]
}
