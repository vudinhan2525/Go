package main

import (
	"fmt"
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func minOperations(boxes string) []int {
	n := len(boxes)
	ans := []int{}
	runes := []rune(boxes)
	for i := 0; i < n; i++ {
		sum := 0
		for j := 0; j < n; j++ {
			if runes[j] == '1' {
				sum += (Abs(j - i))
			}
		}
		ans = append(ans, sum)
	}
	return ans
}
func main() {
	ans := minOperations("10010100")
	fmt.Println(ans)
}
