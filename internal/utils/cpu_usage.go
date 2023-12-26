// From: https://stackoverflow.com/a/31030753

package utils

//#include <time.h>
import "C"
import "time"

var startTime = time.Now()
var startTicks = C.clock()

// CPUUsagePercent returns the CPU usage percent.
func CPUUsagePercent() float64 {
	clockSeconds := float64(C.clock()-startTicks) / float64(C.CLOCKS_PER_SEC)
	realSeconds := time.Since(startTime).Seconds()
	return clockSeconds / realSeconds * 100
}
