package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	cpuCores = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "cpu_cores",
			Help:      "Number of CPU cores",
		},
		[]string{"vserver", "nickname"})
	memory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "memory_bytes",
			Help:      "Amount of Memory in Bytes",
		},
		[]string{"vserver", "nickname"})
	monthlyTrafficIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_in_bytes",
			Help:      "Monthly traffic incoming in Bytes (only gigabyte-level resolution)",
		},
		[]string{"vserver", "nickname", "month", "year", "mac"})
	monthlyTrafficOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_out_bytes",
			Help:      "Monthly traffic outgoing in Bytes (only gigabyte-level resolution)",
		},
		[]string{"vserver", "nickname", "month", "year", "mac"})
	monthlyTrafficTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "monthlytraffic_total_bytes",
			Help:      "Total monthly traffic in Bytes (only gigabyte-level resolution)",
		},
		[]string{"vserver", "nickname", "month", "year", "mac"})
	serverStartTimeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "server_start_time_seconds",
			Help:      "Start time of the vserver in seconds (only minute-level resolution)",
		},
		[]string{"vserver", "nickname"})
	serverIpInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "ip_info",
			Help:      "IPs assigned to this server",
		},
		[]string{"vserver", "nickname", "mac", "ip", "type"})
	ifaceThrottled = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "interface_throttled",
			Help:      "Interface's traffic is throttled (1) or not (0)",
		},
		[]string{"vserver", "nickname", "mac"})
	serverStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "server_status",
			Help:      "Online (1) / Offline (0) status",
		},
		[]string{"vserver", "nickname", "status"})
	rescueActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "rescue_active",
			Help:      "Rescue system active (1) / inactive (0)",
		},
		[]string{"vserver", "nickname"})
	rebootRecommended = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "reboot_recommended",
			Help:      "Reboot recommended (1) / not recommended (0)",
		},
		[]string{"vserver", "nickname"})
	diskCapacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_capacity_bytes",
			Help:      "Available storage space in Bytes",
		},
		[]string{"vserver", "nickname", "driver", "name"})
	diskUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_used_bytes",
			Help:      "Used storage space in Bytes",
		},
		[]string{"vserver", "nickname", "driver", "name"})
	diskOptimization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scp",
			Name:      "disk_optimization",
			Help:      "Optimization recommended (1) / not recommended (0)",
		},
		[]string{"vserver", "nickname"})
)

func init() {
	prometheus.MustRegister(cpuCores)
	prometheus.MustRegister(memory)
	prometheus.MustRegister(monthlyTrafficIn)
	prometheus.MustRegister(monthlyTrafficOut)
	prometheus.MustRegister(monthlyTrafficTotal)
	prometheus.MustRegister(serverStartTimeSeconds)
	prometheus.MustRegister(serverIpInfo)
	prometheus.MustRegister(ifaceThrottled)
	prometheus.MustRegister(serverStatus)
	prometheus.MustRegister(rescueActive)
	prometheus.MustRegister(rebootRecommended)
	prometheus.MustRegister(diskCapacity)
	prometheus.MustRegister(diskUsed)
	prometheus.MustRegister(diskOptimization)
}
