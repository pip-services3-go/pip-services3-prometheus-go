package build

import (
	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	"github.com/pip-services3-go/pip-services3-components-go/build"
	cbuild "github.com/pip-services3-go/pip-services3-components-go/build"
	pcount "github.com/pip-services3-go/pip-services3-prometheus-go/count"
	pservices "github.com/pip-services3-go/pip-services3-prometheus-go/services"
)

// DefaultPrometheusFactory creates Prometheus components by their descriptors.
// See: Factory
// See: PrometheusCounters
// See: PrometheusMetricsService
type DefaultPrometheusFactory struct {
	cbuild.Factory
}

// NewDefaultPrometheusFactory are create a new instance of the factory.
func NewDefaultPrometheusFactory() *DefaultPrometheusFactory {
	c := DefaultPrometheusFactory{}
	c.Factory = *build.NewFactory()

	prometheusCountersDescriptor := cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0")
	prometheusMetricsServiceDescriptor := cref.NewDescriptor("pip-services", "metrics-service", "prometheus", "*", "1.0")

	c.RegisterType(prometheusCountersDescriptor, pcount.NewPrometheusCounters)
	c.RegisterType(prometheusMetricsServiceDescriptor, pservices.NewPrometheusMetricsService)
	return &c
}
