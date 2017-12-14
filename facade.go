package logrusprom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	// autohook prometheus on logrus
	var err error
	hook, err = NewPrometheusHook("log_messages", HandlerOpts(
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			ErrorLog:      ToPrometheusLogger(logrus.StandardLogger()),
		},
	))
	if err != nil {
		panic(err)
	}
	logrus.AddHook(hook)
}

var hook *PrometheusHook

func Handler() http.Handler {
	return hook.Handler()
}

func Registry() *prometheus.Registry {
	return hook.Registry()
}

func Collector() prometheus.Collector {
	return hook.Collector()
}

func SetName(metricName string) error {
	return hook.SetName(metricName)
}

type logrusPromLogger struct {
	logger *logrus.Logger
}

func (l logrusPromLogger) Println(v ...interface{}) {
	l.logger.Error(v...)
}

func ToPrometheusLogger(l *logrus.Logger) promhttp.Logger {
	return logrusPromLogger{l}
}
