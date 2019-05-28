package influx

import "os"

// HostnameMetric reports the hostname for the host.
type HostnameMetric struct{}

// InfluxValue reports the hostname for this host
func (hm HostnameMetric) InfluxValue() interface{} {
	hostname, err := os.Hostname()
	if err != nil {
		return "<UNKNOWN HOSTNAME>"
	}
	return hostname
}
