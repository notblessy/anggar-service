package utils

import "encoding/json"

// Dump :nodoc:
func Dump(data interface{}) string {
	dataByte, _ := json.Marshal(data)
	return string(dataByte)
}
