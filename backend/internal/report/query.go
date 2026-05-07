package report

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// timeFilter 构建时间区间过滤条件。
func timeFilter(field string, from, to int64) bson.M {
	return bson.M{
		field: bson.M{
			"$gte": from,
			"$lt":  to,
		},
	}
}

// shanghaiLocation 返回 Asia/Shanghai 时区。
func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}
