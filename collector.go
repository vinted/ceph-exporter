package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

var schema = make(map[string]string)
var cephMetrics = make(map[string]interface{})
var cephDevice = make(map[string]interface{})
var osdSchema = make(map[string]interface{})
var clusterHealth = make(map[string]cephHealthData)
var mutex = sync.RWMutex{}

const (
	GaugeValue = 2
)

type cephCollector struct {
}

func newCephCollector() *cephCollector {
	return &cephCollector{}
}

func (collector *cephCollector) Describe(ch chan<- *prometheus.Desc) {

}

func LoadJson(jsonData string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		log.Error("Error loading json: ", err)
	}
	// Workaround for radosgw client metrics.
	// Removing hostname from metric name.
	// We need only host name, not FQDN.
	name, _ := os.Hostname()
	var re = regexp.MustCompile(`^([^.]+)`)
	key := "client.radosgw." + re.FindString(name)
	if result[key] != nil {
		log.Debug("Key with hostname found, renaming")
		result["client.radosgw"] = result[key]
		delete(result, key)
	}
	return result
}

func GetDeviceType(socketName string) map[string]string {
	var device = make(map[string]string)
	log.Debug("Getting device info for ", socketName)
	if strings.Contains(socketName, "mon.") {
		log.Debug("Device is a monitor")
		var re = regexp.MustCompile(`mon(.[0-9]*)`)
		device["type"] = "ceph_monitor"
		device["name"] = strings.ReplaceAll(re.FindString(socketName), ".", "")
	}
	if strings.Contains(socketName, "osd.") {
		log.Debug("Device is a osd")
		var re = regexp.MustCompile(`osd(.[0-9]*)`)
		device["type"] = "ceph_osd"
		device["name"] = strings.ReplaceAll(re.FindString(socketName), ".", "")
	}
	if strings.Contains(socketName, "radosgw.") {
		log.Debug("Device is a radosgw")
		var re = regexp.MustCompile(`radosgw(.[0-9]*)`)
		device["type"] = "ceph_radosgw"
		device["name"] = strings.ReplaceAll(re.FindString(socketName), ".", "")
	}
	if strings.Contains(socketName, "mgr.") {
		log.Debug("Device is a mgr")
		var re = regexp.MustCompile(`mgr(.[0-9]*)`)
		device["type"] = "ceph_mgr"
		device["name"] = strings.ReplaceAll(re.FindString(socketName), ".", "")
	}
	log.Debug("Device:", device)
	return device
}

func CollectTimer(queryInterval int) {
	tickChan := time.NewTicker(time.Second * time.Duration(queryInterval))
	quit := make(chan struct{})
	for {
		select {
		case <-tickChan.C:
			log.Debug("Collector timer triggered")
			Collector()
		case <-quit:
			return
		}
	}
}

func Collector() {
	var cephDeviceTmp map[string]string
	log.Debug("Collector started")
	sockets := ListCephSockets()
	for _, socket := range sockets {
		cephDeviceTmp = GetDeviceType(socket)
		if cephDeviceTmp["type"] == "" {
			log.Debug("Not a device. Skipping")
			continue
		}
		mutex.Lock()
		cephDevice[socket] = cephDeviceTmp
		osdSchema[socket] = (LoadJson(GetSchema(socket)))
		cephMetrics[socket] = (LoadJson(GetMetrics(socket)))
		if *healthCollector {
			clusterHealth = CephHealthCollector()
		}
		mutex.Unlock()
	}
	log.Debug("Collector stopped")
}

func (collector *cephCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeTime := time.Now()
	log.Debug("Processing HTTP request")
	mutex.RLock()
	for socket, cephMetric := range cephMetrics {
		for metricName, metricData := range cephMetric.(map[string]interface{}) {
			for metricType, metricsValue := range metricData.(map[string]interface{}) {
				metricSchema, ok := osdSchema[socket].(map[string]interface{})[metricName]
				// There's a possibility, that no full schema is yet available when ceph daemon
				// is starting. Thus we should check on that and destroy partial schema.
				if !ok {
					log.Debug("Missing schema for metric, - socket might be starting up: ", socket)
					// Delete partial schema
					delete(schema, socket)
					continue
				}
				metric := metricSchema.(map[string]interface{})[metricType]
				dataType := metric.(map[string]interface{})["type"].(float64)
				metricDescription := metric.(map[string]interface{})["description"].(string)
				normalizedMetricName := CephNormalizeMetricName(metricName)

				// There are metrics with second level of data (SUMs and AVGs)
				if reflect.TypeOf(metricsValue).Kind() == reflect.Map {
					for metricType1, metricsValue1 := range metricsValue.(map[string]interface{}) {
						description := CephPrometheusDesc(cephDevice[socket].(map[string]string)["type"]+"_"+normalizedMetricName+"_"+metricType+"_"+metricType1, metricDescription)
						ch <- prometheus.MustNewConstMetric(description, GetDatatype(dataType), metricsValue1.(float64), cephDevice[socket].(map[string]string)["name"])
					}
				} else {
					description := CephPrometheusDesc(cephDevice[socket].(map[string]string)["type"]+"_"+normalizedMetricName+"_"+metricType, metricDescription)
					ch <- prometheus.MustNewConstMetric(description, GetDatatype(dataType), metricsValue.(float64), cephDevice[socket].(map[string]string)["name"])
				}
			}
		}
	}
	for clusterHealthMetric, clusterHealthData := range clusterHealth {
		description := CephPrometheusDesc(clusterHealthMetric, clusterHealthData.help)
		ch <- prometheus.MustNewConstMetric(description, GetDatatype(clusterHealthData.metricType), clusterHealthData.value, "mon")
	}
	mutex.RUnlock()
	description := prometheus.NewDesc("ceph_exporter_scrape_time", "Duration of a collector scrape", nil, nil)
	ch <- prometheus.MustNewConstMetric(description, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
	log.Debug("HTTP request finished")
}

// Remove resctricted characters from metric name
// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
func CephNormalizeMetricName(metric string) string {
	metric = strings.ReplaceAll(metric, "-", "_")
	metric = strings.ReplaceAll(metric, ".", "_")
	return strings.ReplaceAll(metric, "__", "_")
}

func CephPrometheusDesc(metricName string, description string) *prometheus.Desc {
	return prometheus.NewDesc(metricName, description, []string{"device"}, nil)
}

// Get schema for defined socket. Either query ceph or use stored map if exists.
func GetSchema(socket string) string {
	log.Debug("Searching inmemory schema for: ", socket)
	if len(schema[socket]) > 0 {
		log.Debug("Inmemory schema found")
	} else {
		log.Debug("Inmemory schema missing. Generating schema.")
		cmdOutput, _ := exec.Command("ceph", "--admin-daemon", socket, "perf", "schema").Output()
		schema[socket] = string(cmdOutput)
	}
	return schema[socket]
}

// Get metrics from defined socket
func GetMetrics(socket string) string {
	log.Debug("Getting metrics for ", socket)
	cmdOutput, _ := exec.Command("ceph", "--admin-daemon", socket, "perf", "dump").Output()
	return string(cmdOutput)
}

// Get a list of ceph admin sockets
func ListCephSockets() []string {
	log.Debug("Getting ceph asok list")
	sockets, err := filepath.Glob(*asokPath + "/*.asok")
	if err != nil {
		log.Fatal(err)
	}
	return sockets
}

// Return correct datatype for metric.
// https://docs.ceph.com/docs/master/dev/perf_counters/
func GetDatatype(dataType float64) prometheus.ValueType {
	switch dataType {
	case 2:
		return prometheus.GaugeValue
	case 5:
		return prometheus.GaugeValue
	case 10:
		return prometheus.CounterValue
	}
	return prometheus.GaugeValue
}
