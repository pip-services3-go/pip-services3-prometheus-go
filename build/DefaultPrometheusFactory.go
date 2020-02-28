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
	Descriptor                         *cref.Descriptor
	PrometheusCountersDescriptor       *cref.Descriptor
	PrometheusMetricsServiceDescriptor *cref.Descriptor
}

// NewDefaultPrometheusFactory are create a new instance of the factory.
func NewDefaultPrometheusFactory() *DefaultPrometheusFactory {
	c := DefaultPrometheusFactory{}
	c.Factory = *build.NewFactory()
	c.Descriptor = cref.NewDescriptor("pip-services", "factory", "prometheus", "default", "1.0")
	c.PrometheusCountersDescriptor = cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0")
	c.PrometheusMetricsServiceDescriptor = cref.NewDescriptor("pip-services", "metrics-service", "prometheus", "*", "1.0")

	c.RegisterType(c.PrometheusCountersDescriptor, pcount.NewPrometheusCounters)
	c.RegisterType(c.PrometheusMetricsServiceDescriptor, pservices.NewPrometheusMetricsService)
	return &c
}
