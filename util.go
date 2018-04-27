package executor

import (
	"hash/crc32"
	"strconv"
)

func intToStr(i int) string {
	return strconv.Itoa(i)
}

func strToInt(str string) int {
	i, _ := strconv.Atoi(str)
	return i
}

func getWorkerID(key string, maxWorkers uint) int {
	i := int(crc32.ChecksumIEEE([]byte(key)))
	n := (i % int(maxWorkers)) + 1
	return n
}
