package influx

import "time"

// DurationMetric is a metric that tracks a duration of time and reports it in milliseconds as a float64.
type DurationMetric struct {
	startTime time.Time
	duration  time.Duration
}

// Start starts the time tracking for the DurationMetric
func (dm *DurationMetric) Start() {
	dm.startTime = time.Now()
}

// Stop stops the DurationMetric and records the result
func (dm *DurationMetric) Stop() {
	dm.duration = time.Since(dm.startTime)
}

// InfluxValue reports the value of the metric in milliseconds as a float64. If the DurationMetric has not been stopped, it will automatically be stopped by this call.
func (dm DurationMetric) InfluxValue() interface{} {
	if dm.duration == 0 && !dm.startTime.IsZero() {
		dm.Stop()
	}
	return dm.duration.Seconds() * 1e3
}
