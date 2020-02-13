package main

import (
	"reflect"
	"testing"
)

func TestCephHealthCollector(t *testing.T) {
	health := CephHealthCollector()
	if reflect.TypeOf(health["ceph_cluster_pgs_degraded"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_degraded] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_undersized"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_undersized] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_stuck_degraded"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_stuck_degraded] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_stuck_unclean"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_degraded] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_backfill"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_backfill] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_backfill_toofull"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_backfill_toofull] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_backfill_wait"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_backfill_wait] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_recovery_wait"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_recovery_wait] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_pgs_peering"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_pgs_peering] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_objects_degraded"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_objects_degraded] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_objects_misplaced"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_objects_misplaced] has wrong data")
	}
	if reflect.TypeOf(health["ceph_cluster_noout_flag"]).String() != "main.cephHealthData" {
		t.Errorf("health[ceph_cluster_noout_flag] has wrong data")
	}
}
