package statistics

import (
	"github.com/markusressel/fan2go/internal/fans"
	"github.com/prometheus/client_golang/prometheus"
)

const fanSubsystem = "fan"

type FanCollector struct {
	fans []fans.Fan
	pwm  *prometheus.Desc
	rpm  *prometheus.Desc
}

func NewFanCollector(fans []fans.Fan) *FanCollector {
	return &FanCollector{
		fans: fans,
		pwm: prometheus.NewDesc(prometheus.BuildFQName(namespace, fanSubsystem, "pwm"),
			"Current PWM value of the fan",
			[]string{"id"}, nil,
		),
		rpm: prometheus.NewDesc(prometheus.BuildFQName(namespace, fanSubsystem, "rpm"),
			"Current RPM value of the fan",
			[]string{"id"}, nil,
		),
	}
}

func (collector *FanCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.pwm
	ch <- collector.rpm
}

// Collect implements required collect function for all prometheus collectors
func (collector *FanCollector) Collect(ch chan<- prometheus.Metric) {
	for _, fan := range collector.fans {
		fanId := fan.GetId()

		pwm, _ := fan.GetPwm()
		ch <- prometheus.MustNewConstMetric(collector.pwm, prometheus.GaugeValue, float64(pwm), fanId)

		if fan.Supports(fans.FeatureRpmSensor) {
			rpm, _ := fan.GetRpm()
			ch <- prometheus.MustNewConstMetric(collector.rpm, prometheus.GaugeValue, float64(rpm), fanId)
		}
	}
}
