# Prometheus ceph exporter

**Information**

Exporter queries `ceph admin sockets` (`asok`) and generates detailed metrics for:  
`OSDs`  
`MONs`  
`RGWs`  
`Any other (future and present) instances which support asok sockets`  

Metric names are generated from socket schema.  
Thus it should not depend on `ceph` version and work with all `ceph` releases.  

**Building**

Checkout https://github.com/vinted/ceph-exporter repo.  
Build executable:  

 `go build`

**Using**

Execute ceph-exporter:  

`./ceph-exporter`

By default exporter will bind to port `9353`.  

**Configuration**

Following config parameters are available:  

```
  -asok.path string
    	path to ceph admin socket direcotry (default "/var/run/ceph")
  -log.level string
    	Logging level (default "info")
  -telemetry.addr string
    	host:port for ceph exporter (default ":9353")
