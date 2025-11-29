package lokiparser

import (
	"strconv"
	"time"
)

// parse str to time nano
func ParseStrToTimeNano(ns string) time.Time {
	n, _ := strconv.ParseInt(ns, 10, 64)
	return time.Unix(0, n)
}
