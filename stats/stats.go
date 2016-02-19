package stats

import (
	"expvar"
	"strconv"
)

var (
	appName string
	tags    = make([]string, 0)

	expStats = expvar.NewMap("stats")
)

type floatStringer struct {
	val float64
}

func (fs floatStringer) String() string {
	return strconv.FormatFloat(fs.val, 'E', -1, 64)
}

// Init initializes the stats
func Init(name string) {
	appName = name
	tags = append(tags, name)
}

// Gauge sets the latest value of a key
func Gauge(key string, val float64) {
	expStats.Set(key, floatStringer{val})
}

// Add keeps a counter
func Add(key string, val float64) {
	expStats.AddFloat(key, val)
}

// Histogram tracks a histogram of a group of measurements
func Histogram(key string, val float64) {
}
