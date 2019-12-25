package executor

import (
	"strconv"
)

func intToStr(i int) string {
	return strconv.Itoa(i)
}

func strToInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}
