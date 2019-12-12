package main

import (
  "encoding/json"
  "github.com/prometheus/client_golang/prometheus"
  log "github.com/Sirupsen/logrus"
  "os"
  "os/exec"
  "path/filepath"
  "reflect"
  "regexp"
  "strings"
  "time"
)

var schema = make(map[string]string)

type cephCollector struct {

}

func newCephCollector() *cephCollector {
	return &cephCollector{}
}

func (collector *cephCollector) Describe(ch chan<- *prometheus.Desc) {

}

func LoadJson(jsonData string) map[string]interface{} {
  var result map[string]interface{}
  json.Unmarshal([]byte(jsonData), &result)
  // Workaround for radosgw client metrics.
  // Removing hostname from metric name.
  // We need only host name, not FQDN.
  name, _ := os.Hostname()
  var re = regexp.MustCompile(`^([^.]+)`)
  key := "client.radosgw." + re.FindString(name)
  if result[key] != nil {
    log.Debug("Key with hostname found, renaming")
    result["client.radosgw"] = result[key]
    delete (result, key)
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
  log.Debug("Device:", device)
  return device
}

func (collector *cephCollector) Collect(ch chan<- prometheus.Metric) {

  scrapeTime := time.Now()
  var osdSchema map[string]interface{}
  var cephMetrics map[string]interface{}
  var cephDevice map[string]string
  var dataType float64
  var metricDescription string

  log.Debug("Processing request")
  sockets := ListCephSockets()
  for _, socket := range sockets {
    cephDevice = GetDeviceType(socket)
    if cephDevice["type"] == "" {
      log.Debug("Not a device. Skipping")
      continue
    }
    cephMetrics = (LoadJson(GetMetrics(socket)))
    osdSchema = (LoadJson(GetSchema(socket)))
    for metricName, metricData := range cephMetrics {
      for metricType, metricsValue := range metricData.(map[string]interface{}) {
        dataType = osdSchema[metricName].(map[string]interface{})[metricType].(map[string]interface{})["type"].(float64)
        metricDescription = osdSchema[metricName].(map[string]interface{})[metricType].(map[string]interface{})["description"].(string)
        var normalizedMetricName string

        normalizedMetricName = CephNormalizeMetricName(metricName)
        // There are metrics with second level of data (SUMs and AVGs)
        if reflect.TypeOf(metricsValue).Kind() == reflect.Map {
          for metricType1, metricsValue1 := range metricsValue.(map[string]interface{}) {
            description := CephPrometheusDesc(cephDevice["type"] + "_" + normalizedMetricName + "_" + metricType + "_" + metricType1, metricDescription)
            ch <- prometheus.MustNewConstMetric(description, GetDatatype(dataType), metricsValue1.(float64), cephDevice["name"])
          }
        } else {
          description := CephPrometheusDesc(cephDevice["type"] + "_" + normalizedMetricName + "_" + metricType, metricDescription)
          ch <- prometheus.MustNewConstMetric(description, GetDatatype(dataType), metricsValue.(float64), cephDevice["name"])
        }
      }
    }
  }
  description := prometheus.NewDesc("ceph_exporter_scrape_time", "Duration of a collector scrape", nil, nil)
  ch <- prometheus.MustNewConstMetric(description, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
  log.Debug("Finished")
}

// Remove resctricted characters from metric name
// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
func CephNormalizeMetricName(metric string) string {
  metric = strings.ReplaceAll(metric, "-", "_")
  metric = strings.ReplaceAll(metric, ".", "_")
  return strings.ReplaceAll(metric, "__", "_")
}

func CephPrometheusDesc(metricName string, description string) *prometheus.Desc{
  return prometheus.NewDesc(metricName, description, []string{"device"}, nil,)
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
