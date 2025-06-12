package types

import (
	"strconv"
	"time"
)

type DataMap map[string]interface{}

func (d DataMap) GetLastModifiedTimeStamp() time.Time {
	timestamp, ok := d["_last_modified"]
	if !ok {
		return time.Unix(0, 0).UTC()
	}
	timeInt, _ := strconv.ParseInt(timestamp.(string), 10, 64)
	return time.UnixMicro(timeInt).UTC()
}
