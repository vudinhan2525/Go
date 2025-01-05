package main

import (
	"fmt"
	"math/big"
)

var curAns = []*big.Int{}

func backtrack(i int, s string) bool {
	if i == len(s) && len(curAns) >= 3 && curAns[len(curAns)-2].Add(curAns[len(curAns)-2], curAns[len(curAns)-3]).Cmp(curAns[len(curAns)-1]) == 0 {
		return true
	}
	for j := i + 1; j <= len(s); j++ {
		char := s[i:j]
		if char[0] == '0' && len(char) > 1 {
			break
		}
		vl := new(big.Int)
		vl, ok := vl.SetString(char, 10)
		if !ok {
			fmt.Println("Error converting number")
			break
		}
		if len(curAns) >= 2 && curAns[len(curAns)-2].Add(curAns[len(curAns)-2], curAns[len(curAns)-1]).Cmp(vl) < 0 {
			break
		}
		if len(curAns) < 2 || curAns[len(curAns)-2].Add(curAns[len(curAns)-2], curAns[len(curAns)-1]).Cmp(vl) == 0 {
			curAns = append(curAns, vl)
			fmt.Println("Current sequence:", curAns)
			if backtrack(j, s) {
				return true
			}
			curAns = curAns[:len(curAns)-1]
		}
	}

	return false
}
func isAdditiveNumber(num string) bool {
	return backtrack(0, num)
}
func main() {
	if isAdditiveNumber("11235813213455890144") {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}
}
