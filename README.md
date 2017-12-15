# Logrusprom

A [logrus](https://github.com/sirupsen/logrus) hook to export total numbers of log message by their level and by 
type for [Prometheus](https://prometheus.io/).

## Ouput example

```promql
# HELP log_messages Total number of log_messages .
# TYPE log_messages counter
log_messages{level="debug",type="untyped"} 0
log_messages{level="error",type="untyped"} 1
log_messages{level="fatal",type="untyped"} 0
log_messages{level="info",type="untyped"} 1
log_messages{level="panic",type="untyped"} 0
log_messages{level="warning",type="WarningOnSomething"} 1
log_messages{level="warning",type="untyped"} 0
```

## Install

Run `go get github.com/ArthurHlt/logrusprom`

## Usage

### Simplest

```go
package main

import (
	"github.com/ArthurHlt/logrusprom"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
    // When importing logrusprom, a hook is generated and added to logrus automatically
    // You only need to use logrus as you normally do
	log.Info("info")
	log.Error("error")
	// by adding a field in form of logrusprom.ErrorTypeKey you can set a type to your metric
	// this is useful for alerting on particular error type
	log.WithField(logrusprom.TypeKey, "WarningOnSomething").Warn("warning")
	
	// add the handler to retrieve metrics
	http.ListenAndServe(":8080", logrusprom.Handler())
	// this give the output we gave as example
}
```

**Tips**: 
- You can rename your metric name by doing `logrusprom.SetName("my_custom_metric_name")` (By default name is: `log_messages`)
- You can add custom labels to your metrics: `logrusprom.SetLabels(map[string]string{"mylabel": "labelvalue"})`

### Using in your own prometheus registry

```go
package main

import (
	"github.com/ArthurHlt/logrusprom"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log.Info("info")
	log.Error("error")
	log.WithField(logrusprom.TypeKey, "WarningOnSomething").Warn("warning")

    // Add the collector in your registry (here we use the default one)
	prometheus.MustRegister(logrusprom.Collector())
	http.ListenAndServe(":8080", promhttp.Handler())
}
```

### By creating hook yourself (useful when not using default logrus)

```go
package main

import (
	"github.com/ArthurHlt/logrusprom"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	myLogger := logrus.New()
	hook, err := logrusprom.NewPrometheusHook("log_messages", logrusprom.HandlerOpts(
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			ErrorLog:      logrusprom.ToPrometheusLogger(myLogger),
		},
	))
	if err != nil {
		panic(err)
	}

	myLogger.AddHook(hook)

	myLogger.Info("info")
	myLogger.Error("error")
	myLogger.WithField(logrusprom.TypeKey, "WarningOnSomething").Warn("warning")

	http.ListenAndServe(":8080", hook.Handler())
}
```
