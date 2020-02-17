package count

import (
	"bytes"
	"net/http"
	"os"
	"time"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	cerr "github.com/pip-services3-go/pip-services3-commons-go/errors"
	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	ccount "github.com/pip-services3-go/pip-services3-components-go/count"
	cinfo "github.com/pip-services3-go/pip-services3-components-go/info"
	clog "github.com/pip-services3-go/pip-services3-components-go/log"
	rpcconnect "github.com/pip-services3-go/pip-services3-rpc-go/connect"
)

/*
Performance counters that send their metrics to Prometheus service.

The component is normally used in passive mode conjunction with PrometheusMetricsService.
Alternatively when connection parameters are set it can push metrics to Prometheus PushGateway.

 Configuration parameters

- connection(s):
  - discovery_key:         (optional) a key to retrieve the connection from connect.idiscovery.html IDiscovery
  - protocol:              connection protocol: http or https
  - host:                  host name or IP address
  - port:                  port number
  - uri:                   resource URI or connection string with all parameters in it
- options:
  - retries:               number of retries (default: 3)
  - connect_timeout:       connection timeout in milliseconds (default: 10 sec)
  - timeout:               invocation timeout in milliseconds (default: 10 sec)

 References

- *:logger:*:*:1.0         (optional) ILogger components to pass log messages
- *:counters:*:*:1.0         (optional) ICounters components to pass collected measurements
- *:discovery:*:*:1.0        (optional)  IDiscovery services to resolve connection

See sRestService
See  CommandableHttpService

 Example

    let counters = new PrometheusCounters();
    counters.configure(ConfigParams.fromTuples(
        "connection.protocol", "http",
        "connection.host", "localhost",
        "connection.port", 8080
    ));

    counters.open("123", (err) => {
        ...
    });

    counters.increment("mycomponent.mymethod.calls");
    let timing = counters.beginTiming("mycomponent.mymethod.exec_time");
    try {
        ...
    } finally {
        timing.endTiming();
    }

    counters.dump();
*/
//implements IReferenceable, IOpenable {
type PrometheusCounters struct {
	*ccount.CachedCounters
	logger             *clog.CompositeLogger
	connectionResolver *rpcconnect.HttpConnectionResolver
	opened             bool
	source             string
	instance           string
	client             *http.Client
	requestRoute       string
	timeout            int
	retries            int
	connectTimeout     int
	uri                string
}

/*
Creates a new instance of the performance counters.
*/
func NewPrometheusCounters() *PrometheusCounters {
	// super();
	pc := PrometheusCounters{}
	pc.CachedCounters = ccount.InheritCacheCounters(&pc)
	pc.logger = clog.NewCompositeLogger()
	pc.connectionResolver = rpcconnect.NewHttpConnectionResolver()
	pc.opened = false
	pc.timeout = 10000
	pc.retries = 3
	pc.connectTimeout = 10000
	return &pc
}

/*
Configures component by passing configuration parameters.

- config    configuration parameters to be set.
*/
func (c *PrometheusCounters) Configure(config *cconf.ConfigParams) {

	c.CachedCounters.Configure(config)

	c.connectionResolver.Configure(config)
	c.source = config.GetAsStringWithDefault("source", c.source)
	c.instance = config.GetAsStringWithDefault("instance", c.instance)
	c.retries = config.GetAsIntegerWithDefault("options.retries", c.retries)
	c.connectTimeout = config.GetAsIntegerWithDefault("options.connectTimeout", c.connectTimeout)
	c.timeout = config.GetAsIntegerWithDefault("options.timeout", c.timeout)
}

/*
Sets references to dependent components.
 *
- references 	references to locate the component dependencies.
*/
func (c *PrometheusCounters) SetReferences(references cref.IReferences) {
	c.logger.SetReferences(references)
	c.connectionResolver.SetReferences(references)
	ref := references.GetOneOptional(
		cref.NewDescriptor("pip-services", "context-info", "default", "*", "1.0"))
	contextInfo, _ := ref.(*cinfo.ContextInfo)
	if contextInfo != nil && c.source == "" {
		c.source = contextInfo.Name
	}
	if contextInfo != nil && c.instance == "" {
		c.instance = contextInfo.ContextId
	}
}

/*
Checks if the component is opened.

Returns true if the component has been opened and false otherwise.
*/
func (c *PrometheusCounters) IsOpen() bool {
	return c.opened
}

/*
Opens the component.

- correlationId 	(optional) transaction id to trace execution through call chain.
- callback 			callback function that receives error or nil no errors occured.
*/
func (c *PrometheusCounters) Open(correlationId string) (err error) {
	if c.opened {
		return nil
	}

	c.opened = true
	connection, _, err := c.connectionResolver.Resolve(correlationId)

	if err != nil {
		c.client = nil
		c.logger.Warn(correlationId, "Connection to Prometheus server is not configured: "+err.Error())
		return nil
	}

	c.uri = connection.Uri()

	job := c.source
	if job == "" {
		job = "unknown"
	}

	instance := c.instance
	if instance == "" {
		host, _ := os.Hostname()
		instance = host
	}
	c.requestRoute = "/metrics/job/" + job + "/instance/" + instance

	localClient := http.Client{}
	localClient.Timeout = (time.Duration)(c.timeout) * time.Millisecond
	c.client = &localClient
	if c.client == nil {
		ex := cerr.NewConnectionError(correlationId, "CANNOT_CONNECT", "Connection to REST service failed").WithDetails("url", c.uri)
		return ex
	}

	return nil
}

/*
Closes component and frees used resources.

- correlationId 	(optional) transaction id to trace execution through call chain.
- callback 			callback function that receives error or nil no errors occured.
*/
func (c *PrometheusCounters) Close(correlationId string) error {
	c.opened = false
	c.client = nil
	c.requestRoute = ""
	return nil
}

/*
Saves the current counters measurements.

- counters      current counters measurements to be saves.
*/
func (c *PrometheusCounters) Save(counters []*ccount.Counter) (err error) {
	if c.client == nil {
		return nil
	}

	url := c.uri + c.requestRoute
	body := PrometheusCounterConverter.ToString(counters, "", "")

	req, reqErr := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(body)))
	if reqErr != nil {
		err = cerr.NewUnknownError("PrometheusCounters", "UNSUPPORTED_METHOD", "Method is not supported by REST client").WithDetails("verb", "PUT").WithCause(reqErr)
		return err
	}
	// Set headers
	req.Header.Set("Accept", "text/html")
	retries := c.retries
	var resp *http.Response
	var respErr error

	for retries > 0 {
		// Try send request
		resp, respErr = c.client.Do(req)
		if respErr != nil {

			retries--
			if retries == 0 {
				err = cerr.NewUnknownError("PrometheusCounters", "COMMUNICATION_ERROR", "Unknown communication problem on REST client").WithCause(respErr)
				return err
			}
			continue

		}
		break
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 204 && resp.StatusCode < 300 {
			return nil
		}
		c.logger.Error("prometheus-counters", respErr, "Failed to push metrics to prometheus")
	}

	return respErr
}
