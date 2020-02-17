package test_services

import (
	"io/ioutil"
	"net/http"
	"testing"

	cconf "github.com/pip-services3-go/pip-services3-commons-go/config"
	cref "github.com/pip-services3-go/pip-services3-commons-go/refer"
	cinfo "github.com/pip-services3-go/pip-services3-components-go/info"
	pcount "github.com/pip-services3-go/pip-services3-prometheus-go/count"
	pservice "github.com/pip-services3-go/pip-services3-prometheus-go/services"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMetricsService(t *testing.T) {
	var service *pservice.PrometheusMetricsService
	var counters *pcount.PrometheusCounters

	var restConfig = cconf.NewConfigParamsFromTuples(
		"connection.protocol", "http",
		"connection.host", "localhost",
		"connection.port", "3000",
	)

	service = pservice.NewPrometheusMetricsService()
	service.Configure(restConfig)

	counters = pcount.NewPrometheusCounters()

	contextInfo := cinfo.NewContextInfo()
	contextInfo.Name = "Test"
	contextInfo.Description = "This is a test container"

	references := cref.NewReferencesFromTuples(
		cref.NewDescriptor("pip-services", "context-info", "default", "default", "1.0"), contextInfo,
		cref.NewDescriptor("pip-services", "counters", "prometheus", "default", "1.0"), counters,
		cref.NewDescriptor("pip-services", "metrics-service", "prometheus", "default", "1.0"), service,
	)
	counters.SetReferences(references)
	service.SetReferences(references)

	opnErr := counters.Open("")
	if opnErr == nil {
		service.Open("")
	}

	defer service.Close("")
	defer counters.Close("")

	var url = "http://localhost:3000"

	counters.IncrementOne("test.counter1")
	counters.Stats("test.counter2", 2)
	counters.Last("test.counter3", 3)
	counters.TimestampNow("test.counter4")

	getRes, getErr := http.Get(url + "/metrics")
	assert.Nil(t, getErr)
	assert.NotNil(t, getRes)
	assert.True(t, getRes.StatusCode < 400)
	body, _ := ioutil.ReadAll(getRes.Body)
	assert.True(t, len(body) > 0)

	for {
	}
}
