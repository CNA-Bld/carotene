package utils

import (
	"strconv"
	"strings"
	"time"
)

func TimeFromFileName(fileName string) (time.Time, error) {
	t, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimSuffix(fileName, "Q.msgpack"), "R.msgpack"), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(t), nil
}
