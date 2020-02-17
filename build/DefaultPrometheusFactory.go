package build

import (
	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	"github.com/pip-services3-go/pip-services3-components-go/build"
	cbuild "github.com/pip-services3-go/pip-services3-components-go/build"
	pcount "github.com/pip-services3-go/pip-services3-prometheus-go/count"
	pservices "github.com/pip-services3-go/pip-services3-prometheus-go/services"
)

// import { PrometheusCounters } from '../count/PrometheusCounters';
// import { PrometheusMetricsService } from '../services/PrometheusMetricsService';

/**
 * Creates Prometheus components by their descriptors.
 *
 * @see [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/classes/build.factory.html Factory]]
 * @see [[PrometheusCounters]]
 * @see [[PrometheusMetricsService]]
 */
type DefaultPrometheusFactory struct {
	cbuild.Factory
	Descriptor                         *cref.Descriptor
	PrometheusCountersDescriptor       *cref.Descriptor
	PrometheusMetricsServiceDescriptor *cref.Descriptor
}

/**
 * Create a new instance of the factory.
 */
func NewDefaultPrometheusFactory() *DefaultPrometheusFactory {
	dpf := DefaultPrometheusFactory{}
	dpf.Factory = *build.NewFactory()
	dpf.Descriptor = cref.NewDescriptor("pip-services", "factory", "prometheus", "default", "1.0")
	dpf.PrometheusCountersDescriptor = cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0")
	dpf.PrometheusMetricsServiceDescriptor = cref.NewDescriptor("pip-services", "metrics-service", "prometheus", "*", "1.0")

	dpf.RegisterType(dpf.PrometheusCountersDescriptor, pcount.NewPrometheusCounters)
	dpf.RegisterType(dpf.PrometheusMetricsServiceDescriptor, pservices.NewPrometheusMetricsService)
	return &dpf
}
