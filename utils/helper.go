package utils

import (
	"encoding/json"
	"strconv"
)

// Dump :nodoc:
func Dump(data interface{}) string {
	dataByte, _ := json.Marshal(data)
	return string(dataByte)
}

func ParseID(n string) int64 {
	id, _ := strconv.ParseInt(n, 10, 64)
	return id
}
