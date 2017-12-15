package logrusprom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strings"
)

const (
	TypeKey     = "type"
	defaultType = "untyped"
)

type optSetter func(f *PrometheusHook)

type PrometheusHook struct {
	counterVec  *prometheus.CounterVec
	promReg     *prometheus.Registry
	handlerOpts promhttp.HandlerOpts
	labels      map[string]string
	metricName  string
}

func NewPrometheusHook(metricName string, setters ...optSetter) (*PrometheusHook, error) {
	hook := &PrometheusHook{
		labels:     make(map[string]string),
		metricName: metricName,
	}
	for _, s := range setters {
		s(hook)
	}
	err := hook.initCounter()
	if err != nil {
		return nil, err
	}
	return hook, nil
}

func (h PrometheusHook) Fire(entry *logrus.Entry) error {
	errType := defaultType
	if errTypeI, ok := entry.Data[TypeKey]; ok {
		errType = sanitizeName(fmt.Sprint(errTypeI))
	}
	labelValues := []string{entry.Level.String(), errType}
	labelValues = append(labelValues, valuesOrderFromMap(h.labels)...)
	h.counterVec.WithLabelValues(labelValues...).Inc()
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
	h.metricName = metricName
	return h.initCounter()
}

func (h *PrometheusHook) SetLabels(labels map[string]string) error {
	h.labels = labels
	return h.initCounter()
}

func (h *PrometheusHook) initCounter() error {
	h.promReg = prometheus.NewRegistry()

	labelKeys := []string{"level", TypeKey}
	labelKeys = append(labelKeys, keysOrderFromMap(h.labels)...)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: sanitizeName(h.metricName),
		Help: fmt.Sprintf("Total number of %s .", h.metricName),
	}, labelKeys)

	h.counterVec = counterVec
	return h.promReg.Register(h.counterVec)
}
func sanitizeName(s string) string {
	return strings.Replace(strings.TrimSpace(s), " ", "_", -1)
}
func valuesOrderFromMap(m map[string]string) []string {
	values := make([]string, 0)
	for _, k := range keysOrderFromMap(m) {
		values = append(values, m[k])
	}
	return values
}
func keysOrderFromMap(m map[string]string) []string {
	keys := make([]string, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func HandlerOpts(opts promhttp.HandlerOpts) optSetter {
	return func(h *PrometheusHook) {
		h.handlerOpts = opts
	}
}
func AddLabels(labels map[string]string) optSetter {
	return func(h *PrometheusHook) {
		h.labels = labels
	}
}
