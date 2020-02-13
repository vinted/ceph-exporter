package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"strconv"
)

type cephHealthStats struct {
	Health struct {
		Summary       interface{} `json:"summary"`
		OverallStatus string      `json:"overall_status"`
		Status        string      `json:"status"`
		Checks        interface{} `json:"checks"`
	} `json:"health"`
	PGMap struct {
		NumPGs                  float64 `json:"num_pgs"`
		WriteOpPerSec           float64 `json:"write_op_per_sec"`
		ReadOpPerSec            float64 `json:"read_op_per_sec"`
		WriteBytePerSec         float64 `json:"write_bytes_sec"`
		ReadBytePerSec          float64 `json:"read_bytes_sec"`
		RecoveringObjectsPerSec float64 `json:"recovering_objects_per_sec"`
		RecoveringBytePerSec    float64 `json:"recovering_bytes_per_sec"`
		RecoveringKeysPerSec    float64 `json:"recovering_keys_per_sec"`
		CacheFlushBytePerSec    float64 `json:"flush_bytes_sec"`
		CacheEvictBytePerSec    float64 `json:"evict_bytes_sec"`
		CachePromoteOpPerSec    float64 `json:"promote_op_per_sec"`
		DegradedObjects         float64 `json:"degraded_objects"`
		MisplacedObjects        float64 `json:"misplaced_objects"`
		PGsByState              []struct {
			Count  float64 `json:"count"`
			States string  `json:"state_name"`
		} `json:"pgs_by_state"`
	} `json:"pgmap"`
}

type cephHealthData struct {
	value      float64
	metricType float64
	help       string
}

func CephHealthCollector() map[string]cephHealthData {
	stats := &cephHealthStats{}
	if err := json.Unmarshal(CephHealthCommand(), stats); err != nil {
		log.Debug(err)
	}
	//var healthData = make(map[string]interface{})
	healthData := make(map[string]cephHealthData)
	var healthString string
	var cephHealthStatus float64

	switch stats.Health.Status {
	case "HEALTH_OK":
		cephHealthStatus = 0
	case "HEALTH_WARN":
		cephHealthStatus = 1
	case "HEALTH_ERR":
		cephHealthStatus = 2
	default:
		cephHealthStatus = 2
	}

	switch stats.Health.OverallStatus {
	case "HEALTH_OK":
		cephHealthStatus = 0
	case "HEALTH_WARN":
		cephHealthStatus = 1
	case "HEALTH_ERR":
		cephHealthStatus = 2
	default:
		cephHealthStatus = 2
	}

	healthData["ceph_cluster_pgs_degraded"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of degraded PGs"}
	healthData["ceph_cluster_pgs_undersized"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of undersized PGs"}
	healthData["ceph_cluster_pgs_stuck_degraded"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of stuck degraded PGs"}
	healthData["ceph_cluster_pgs_stuck_unclean"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of stuck unclean PGs"}
	healthData["ceph_cluster_pgs_backfill"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of PGs backfilling"}
	healthData["ceph_cluster_pgs_backfill_toofull"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of PGs too full"}
	healthData["ceph_cluster_pgs_backfill_wait"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of PGs waiting to backfill"}
	healthData["ceph_cluster_pgs_recovery_wait"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of PGs waiting for recovery"}
	healthData["ceph_cluster_pgs_peering"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of peering PGs"}
	healthData["ceph_cluster_objects_degraded"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of degraded objects in a cluster"}
	healthData["ceph_cluster_objects_misplaced"] = cephHealthData{value: 0, metricType: GaugeValue, help: "Number of mislpaced objects in a cluster"}

	// Check if using new Nautilus status data structure.
	if stats.Health.Checks != nil {
		log.Debug("Using new status data structure")
		// Convert whole Health.Checks interface to string.
		healthString = fmt.Sprintf("%v", stats.Health.Checks)
	} else {
		log.Debug("Using old status data structure")
		// Convert whole Health.Checks interface to string.
		healthString = fmt.Sprintf("%v", stats.Health.Summary)
	}

	// Do regex match against status string and build metrics
	re := regexp.MustCompile(`([\d]+) pgs degraded`)
	result := re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_degraded"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_degraded"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs undersized`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_undersized"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_undersized"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs stuck degraded`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_stuck_degraded"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_stuck_degraded"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs stuck unclean`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_stuck_unclean"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_stuck_unclean"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs backfilling`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_backfill"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_backfill"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs backfill_toofull`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_backfill_toofull"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_backfill_toofull"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs backfill_wait`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_backfill_wait"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_backfill_wait"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs recovery_wait`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_recovery_wait"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_recovery_wait"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+) pgs peering`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 2 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_pgs_peering"]
			tmp.value = float64(val)
			healthData["ceph_cluster_pgs_peering"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+)/([\d]+) objects degraded`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 3 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_objects_degraded"]
			tmp.value = float64(val)
			healthData["ceph_cluster_objects_degraded"] = tmp
		}
	}
	re = regexp.MustCompile(`([\d]+)/([\d]+) objects misplaced`)
	result = re.FindStringSubmatch(healthString)
	if len(result) == 3 {
		val, err := strconv.Atoi(result[1])
		if err != nil {
			log.Debug(err)
		} else {
			var tmp = healthData["ceph_cluster_objects_misplaced"]
			tmp.value = float64(val)
			healthData["ceph_cluster_objects_misplaced"] = tmp
		}
	}

	healthData["ceph_cluster_health_status"] = cephHealthData{value: cephHealthStatus, metricType: GaugeValue, help: "Ceph cluster health status (ok:0, warning:1, error:2)"}
	return healthData
}

func CephHealthCommand() []byte {
	log.Debug("Running ceph status")
	cmdOutput, _ := exec.Command("ceph", "-c", *cephConfigFile, "status", "-f", "json").Output()
	return cmdOutput
}
