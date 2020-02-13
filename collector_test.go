package main

import "testing"
import "os"
import "fmt"
import "regexp"
import "github.com/prometheus/client_golang/prometheus"

func TestGetDatatype(t *testing.T) {
	dataType := GetDatatype(2)
	if dataType != prometheus.GaugeValue {
		t.Errorf("Dataype 2 should be GaugeValue")
	}
	dataType = GetDatatype(5)
	if dataType != prometheus.GaugeValue {
		t.Errorf("Dataype 5 should be GaugeValue")
	}
	dataType = GetDatatype(10)
	if dataType != prometheus.CounterValue {
		t.Errorf("Dataype 10 should be CounterValue")
	}
	dataType = GetDatatype(0)
	if dataType != prometheus.GaugeValue {
		t.Errorf("Default datatype should be GaugeValue")
	}
}

func TestCephNormalizeMetricName(t *testing.T) {
	value := CephNormalizeMetricName("test-string.1__")
	if value != "test_string_1_" {
		t.Errorf("CephNormalizeMetricName failed. Got: %s", value)
	}
}

func TestGetDeviceType(t *testing.T) {
	value := GetDeviceType("/var/run/ceph/ceph-cluster-osd.1.asok")
	if value["type"] != "ceph_osd" || value["name"] != "osd1" {
		t.Errorf("GetDeviceType failed. Got: type: %s, name: %s", value["type"], value["name"])
	}
	value = GetDeviceType("/var/run/ceph/ceph-cluster-mon.test-ceph-mon1.asok")
	if value["type"] != "ceph_monitor" || value["name"] != "mon" {
		t.Errorf("GetDeviceType failed. Got: type: %s, name: %s", value["type"], value["name"])
	}
	value = GetDeviceType("/var/run/ceph/ceph-cluster-client.radosgw.test-ceph-osd1.asok")
	if value["type"] != "ceph_radosgw" || value["name"] != "radosgw" {
		t.Errorf("GetDeviceType failed. Got: type: %s, name: %s", value["type"], value["name"])
	}
	value = GetDeviceType("/var/run/ceph/ceph-cluster-client.radosgw.test-ceph-mgr.asok")
	if value["type"] != "ceph_mgr" || value["name"] != "mgr" {
		t.Errorf("GetDeviceType failed. Got: type: %s, name: %s", value["type"], value["name"])
	}
}

func TestLoadJson(t *testing.T) {
	schema := string(`{
      "client.radosgw.%s": {
        "req": {
          "type": 10,
          "description": "Requests",
          "nick": ""
        }
      }
    }`)
	hostname, _ := os.Hostname()
	re := regexp.MustCompile(`^([^.]+)`)
	hostname = re.FindString(hostname)
	schema = fmt.Sprintf(schema, hostname)
	schemaMap := LoadJson(schema)
	result := schemaMap["client.radosgw"].(map[string]interface{})["req"].(map[string]interface{})["type"].(float64)
	if int(result) != 10 {
		t.Errorf("LoadJson failed. Got: %v, needed: 10", result)
	}
}
