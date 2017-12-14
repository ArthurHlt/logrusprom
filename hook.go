package logrusprom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const (
	ErrorTypeKey   = "error_type"
	defaultErrType = "untyped"
)

type optSetter func(f *PrometheusHook)

type PrometheusHook struct {
	counterVec  *prometheus.CounterVec
	promReg     *prometheus.Registry
	handlerOpts promhttp.HandlerOpts
}

func NewPrometheusHook(metricName string, setters ...optSetter) (*PrometheusHook, error) {

	counterVec := createCounterVec(metricName)
	reg := prometheus.NewRegistry()
	err := reg.Register(counterVec)
	if err != nil {
		return nil, err
	}
	hook := &PrometheusHook{
		counterVec: counterVec,
		promReg:    reg,
	}
	for _, s := range setters {
		s(hook)
	}
	return hook, nil
}

func (h PrometheusHook) Fire(entry *logrus.Entry) error {
	errType := defaultErrType
	if errTypeI, ok := entry.Data[ErrorTypeKey]; ok {
		errType = sanitizeName(fmt.Sprint(errTypeI))
	}
	h.counterVec.WithLabelValues(entry.Level.String(), errType).Inc()
	return nil
}

func (PrometheusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h PrometheusHook) Handler() http.Handler {
	return promhttp.HandlerFor(h.promReg, h.handlerOpts)
}

func (h PrometheusHook) Registry() *prometheus.Registry {
	return h.promReg
}

func (h PrometheusHook) Collector() prometheus.Collector {
	return h.counterVec
}

func (h *PrometheusHook) SetName(metricName string) error {
	h.promReg.Unregister(h.counterVec)
	h.counterVec = createCounterVec(metricName)
	return h.promReg.Register(h.counterVec)
}

func createCounterVec(metricName string) *prometheus.CounterVec {
	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: sanitizeName(metricName),
		Help: fmt.Sprintf("Total number of %s .", metricName),
	}, []string{"level", "error_type"})

	for _, level := range logrus.AllLevels {
		counterVec.WithLabelValues(level.String(), defaultErrType)
	}
	return counterVec
}
func sanitizeName(s string) string {
	return strings.Replace(strings.TrimSpace(s), " ", "_", -1)
}

func HandlerOpts(opts promhttp.HandlerOpts) optSetter {
	return func(h *PrometheusHook) {
		h.handlerOpts = opts
	}
}
