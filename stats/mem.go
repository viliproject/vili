package stats

import (
	"runtime"
	"time"
)

// TrackMemStats tracks the number of goroutines and mem stats and refreshes
// every 15 seconds
func TrackMemStats() {
	go func() {
		m := &runtime.MemStats{}
		ticker := time.Tick(15 * time.Second)
		for {
			runtime.ReadMemStats(m)
			Gauge(appName+".goroutines", float64(runtime.NumGoroutine()))
			Gauge(appName+".mem.alloc", float64(m.Alloc))
			Gauge(appName+".mem.sys", float64(m.Sys))
			Gauge(appName+".mem.heapobjects", float64(m.HeapObjects))
			Gauge(appName+".mem.numgc", float64(m.NumGC))
			Histogram(appName+".mem.pausens", float64(m.PauseNs[(m.NumGC+255)%256]))
			<-ticker
		}
	}()

}
