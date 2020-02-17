package services

import (
	"io"
	"net/http"

	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	ccount "github.com/pip-services3-go/pip-services3-components-go/count"
	cinfo "github.com/pip-services3-go/pip-services3-components-go/info"
	pcount "github.com/pip-services3-go/pip-services3-prometheus-go/count"
	rpcservices "github.com/pip-services3-go/pip-services3-rpc-go/services"
)

/*
Service that exposes "/metrics" route for Prometheus to scap performance metrics.

 Configuration parameters

- dependencies:
  - endpoint:              override for HTTP Endpoint dependency
  - prometheus-counters:   override for PrometheusCounters dependency
- connection(s):
  - discovery_key:         (optional) a key to retrieve the connection from IDiscovery
  - protocol:              connection protocol: http or https
  - host:                  host name or IP address
  - port:                  port number
  - uri:                   resource URI or connection string with all parameters in it

 References

- *:logger:*:*:1.0         (optional) [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/interfaces/log.ilogger.html ILogger]] components to pass log messages
- *:counters:*:*:1.0         (optional) [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/interfaces/count.icounters.html ICounters]] components to pass collected measurements
- *:discovery:*:*:1.0        (optional) [[https://rawgit.com/pip-services-node/pip-services3-components-node/master/doc/api/interfaces/connect.idiscovery.html IDiscovery]] services to resolve connection
- *:endpoint:http:*:1.0          (optional) [[https://rawgit.com/pip-services-node/pip-services3-rpc-node/master/doc/api/classes/services.httpendpoint.html HttpEndpoint]] reference to expose REST operation
- *:counters:prometheus:*:1.0    [[PrometheusCounters]] reference to retrieve collected metrics

See [[https://rawgit.com/pip-services-node/pip-services3-rpc-node/master/doc/api/classes/services.restservice.html RestService]]
See [[https://rawgit.com/pip-services-node/pip-services3-rpc-node/master/doc/api/classes/clients.restclient.html RestClient]]

 Example

    let service = new PrometheusMetricsService();
    service.configure(ConfigParams.fromTuples(
        "connection.protocol", "http",
        "connection.host", "localhost",
        "connection.port", 8080
    ));

    service.open("123", (err) => {
       console.log("The Prometheus metrics service is accessible at http://+:8080/metrics");
    });
*/
type PrometheusMetricsService struct {
	*rpcservices.RestService
	cachedCounters *ccount.CachedCounters
	source         string
	instance       string
}

/*
Creates a new instance of c service.
*/
func NewPrometheusMetricsService() *PrometheusMetricsService {
	pms := PrometheusMetricsService{}
	pms.RestService = rpcservices.NewRestService()
	pms.RestService.IRegisterable = &pms
	pms.DependencyResolver.Put("cached-counters", cref.NewDescriptor("pip-services", "counters", "cached", "*", "1.0"))
	pms.DependencyResolver.Put("prometheus-counters", cref.NewDescriptor("pip-services", "counters", "prometheus", "*", "1.0"))
	return &pms
}

/*
Sets references to dependent components.

Return references 	references to locate the component dependencies.
*/
func (c *PrometheusMetricsService) SetReferences(references cref.IReferences) {
	c.RestService.SetReferences(references)

	resolv := c.DependencyResolver.GetOneOptional("prometheus-counters")
	c.cachedCounters = resolv.(*pcount.PrometheusCounters).CachedCounters
	if c.cachedCounters == nil {
		resolv = c.DependencyResolver.GetOneOptional("cached-counters")
		c.cachedCounters = resolv.(*ccount.CachedCounters)
	}
	ref := references.GetOneOptional(
		cref.NewDescriptor("pip-services", "context-info", "default", "*", "1.0"))
	contextInfo := ref.(*cinfo.ContextInfo)

	if contextInfo != nil && c.source == "" {
		c.source = contextInfo.Name
	}
	if contextInfo != nil && c.instance == "" {
		c.instance = contextInfo.ContextId
	}
}

/*
Registers all service routes in HTTP endpoint.
*/
func (c *PrometheusMetricsService) Register() {
	c.RegisterRoute("get", "metrics", nil, func(res http.ResponseWriter, req *http.Request) { c.metrics(res, req) })
}

/*
Handles metrics requests

Return req   an HTTP request
Return res   an HTTP response
*/
func (c *PrometheusMetricsService) metrics(res http.ResponseWriter, req *http.Request) {

	var counters []*ccount.Counter
	if c.cachedCounters != nil {
		counters = c.cachedCounters.GetAll()
	}

	body := pcount.PrometheusCounterConverter.ToString(counters, c.source, c.instance)

	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(200)
	_, wrErr := io.WriteString(res, (string)(body))
	if wrErr != nil {
		c.Logger.Error("PrometheusMetricsService", wrErr, "Can't write response")
	}
}
