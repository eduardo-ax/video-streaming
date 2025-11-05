package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	devices         prometheus.Gauge
	uploads         prometheus.Gauge
	videoUploadTime prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		devices: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "devices_conected",
			Help:      "Number of devices conected.",
		}),
		uploads: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "videos_uploaded",
			Help:      "Number of videos uploaded.",
		}),
		videoUploadTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "myapp",
			Name:      "video_upload_duration_seconds",
			Help:      "Time taken to upload a video file in seconds.",
			Buckets:   prometheus.LinearBuckets(1, 1, 10),
		}),
	}
	reg.MustRegister(m.devices)
	reg.MustRegister(m.uploads)
	reg.MustRegister(m.VideoUploadTime())
	return m
}

func (m *Metrics) UploadsInc() {
	m.uploads.Inc()
}

func (m *Metrics) DevicesInc() {
	m.devices.Inc()
}

func (m *Metrics) VideoUploadTime() prometheus.Histogram {
	return m.videoUploadTime
}
