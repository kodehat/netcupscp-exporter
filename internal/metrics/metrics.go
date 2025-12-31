package metrics

import (
	"github.com/kodehat/netcupscp-exporter/internal/build"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "build_info",
			Help:      "A metric with a constant '1' value labeled by goversion, revision and version from which netcupscp-exporter was built",
		},
		[]string{"goversion", "revision", "version"})
	cpuCores = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "cpu_cores",
			Help:      "Number of CPU cores",
		},
		[]string{"servername", "servernickname"})
	memory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "memory_bytes",
			Help:      "Amount of memory in bytes",
		},
		[]string{"servername", "servernickname"})
	monthlyTrafficIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_in_bytes",
			Help:      "Monthly traffic incoming in bytes (only gigabyte-level resolution)",
		},
		[]string{"servername", "servernickname", "month", "year", "mac"})
	monthlyTrafficOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_out_bytes",
			Help:      "Monthly traffic outgoing in bytes",
		},
		[]string{"servername", "servernickname", "month", "year", "mac"})
	monthlyTrafficTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_total_bytes",
			Help:      "Total monthly traffic in bytes",
		},
		[]string{"servername", "servernickname", "month", "year", "mac"})
	serverStartTimeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "server_start_time_seconds",
			Help:      "Start time of the servername in seconds",
		},
		[]string{"servername", "servernickname"})
	serverIpInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "ip_info",
			Help:      "Ip addresses assigned to this server",
		},
		[]string{"servername", "servernickname", "mac", "ip", "type"})
	ifaceThrottled = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "interface_throttled",
			Help:      "Interface's traffic is throttled (1) or not (0)",
		},
		[]string{"servername", "servernickname", "mac", "status"})
	serverStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "server_status",
			Help:      "Online (1) / Offline (0) status",
		},
		[]string{"servername", "servernickname", "status"})
	rescueActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "rescue_active",
			Help:      "Rescue system active (1) / inactive (0)",
		},
		[]string{"servername", "servernickname", "status"})
	rebootRecommended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "reboot_recommended",
			Help:      "Reboot recommended (1) / not recommended (0)",
		},
		[]string{"servername", "servernickname", "status"})
	diskCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_capacity_bytes",
			Help:      "Available storage space in bytes",
		},
		[]string{"servername", "servernickname", "driver", "name"})
	diskUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_used_bytes",
			Help:      "Used storage space in bytes",
		},
		[]string{"servername", "servernickname", "driver", "name"})
	diskOptimization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_optimization",
			Help:      "Optimization recommended (1) / not recommended (0)",
		},
		[]string{"servername", "servernickname", "status"})
)

func Load() *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		buildInfo,
		cpuCores,
		memory,
		monthlyTrafficIn,
		monthlyTrafficOut,
		monthlyTrafficTotal,
		serverStartTimeSeconds,
		serverIpInfo,
		ifaceThrottled,
		serverStatus,
		rescueActive,
		rebootRecommended,
		diskCapacity,
		diskUsed,
		diskOptimization,
	)

	buildInfo.With(prometheus.Labels{
		"goversion": build.GoVersion,
		"revision":  build.CommitHash,
		"version":   build.Version,
	}).Set(1)

	return registry
}
