package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/client"
	"github.com/kodehat/netcupscp-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
)

func serverBaseLabels(server *client.Server) prometheus.Labels {
	return prometheus.Labels{
		"servername":     *server.Name,
		"servernickname": *server.Nickname,
	}
}

type DefaultMetricsUpdater struct {
	collector collector.ServerCollector
}

var _ MetricsUpdater = DefaultMetricsUpdater{}

func NewDefaultMetricsUpdater(collector collector.ServerCollector) *DefaultMetricsUpdater {
	return &DefaultMetricsUpdater{
		collector: collector,
	}
}

func (mu DefaultMetricsUpdater) UpdateMetrics(context context.Context) error {
	serverInfos, err := mu.collector.CollectServerData(context)
	if err != nil {
		return err
	}
	mu.updateMetricsFromServerInfos(serverInfos)
	return nil
}

func (mu DefaultMetricsUpdater) updateInterfaceMetrics(server *client.Server) {
	baseLabels := serverBaseLabels(server)

	// Update interface specific metrics.
	for _, iface := range *server.ServerLiveInfo.Interfaces {
		trafficLabels := prometheus.Labels{
			"month": fmt.Sprintf("%02d", time.Now().Month()),
			"year":  fmt.Sprintf("%d", time.Now().Year()),
			"mac":   *iface.Mac,
		}

		// Update interface traffic metrics.
		monthlyTrafficIn.With(mergeLabels(baseLabels, trafficLabels)).Set(float64(*iface.RxMonthlyInMiB) * 1024 * 1024)
		monthlyTrafficOut.With(mergeLabels(baseLabels, trafficLabels)).Set(float64(*iface.TxMonthlyInMiB) * 1024 * 1024)
		monthlyTrafficTotal.With(mergeLabels(baseLabels, trafficLabels)).Set(float64(*iface.TxMonthlyInMiB+*iface.RxMonthlyInMiB) * 1024 * 1024)

		// Update interface throttled status.
		ifaceThrottledSatus := INTERFACE_NOT_THROTTLED
		if *iface.TrafficThrottled {
			ifaceThrottledSatus = INTERFACE_THROTTLED
		}
		ifaceThrottled.With(mergeLabels(baseLabels, prometheus.Labels{"mac": *iface.Mac, "status": ifaceThrottledSatus.String()})).Set(float64(ifaceThrottledSatus))

		// Update interface IPv4 info.
		for _, ip := range *iface.Ipv4Addresses {
			ipv4Labels := mergeLabels(baseLabels, prometheus.Labels{"mac": *iface.Mac, "ip": ip, "type": "ipv4"})
			serverIpInfo.With(ipv4Labels).Set(1)
		}
		// Update interface IPv6 info.
		for _, ip := range *iface.Ipv6NetworkPrefixes {
			ipv6Labels := mergeLabels(baseLabels, prometheus.Labels{"mac": *iface.Mac, "ip": ip, "type": "ipv6"})
			serverIpInfo.With(ipv6Labels).Set(1)
		}
	}
}

func (mu DefaultMetricsUpdater) updateDiskMetrics(server *client.Server) {
	baseLabels := serverBaseLabels(server)

	// Update disk optimization status.
	diskOptStatus := DISK_OPTIMIZATION_YES
	if *server.ServerLiveInfo.RequiredStorageOptimization == client.NO {
		diskOptStatus = DISK_OPTIMIZATION_NO
	}
	diskOptimization.With(mergeLabels(baseLabels, prometheus.Labels{"status": diskOptStatus.String()})).Set(float64(diskOptStatus))

	// Update disk capacity and disk usage metrics.
	for _, disk := range *server.ServerLiveInfo.Disks {
		diskLabels := mergeLabels(baseLabels, prometheus.Labels{
			"driver": *disk.Driver,
			"name":   *disk.Dev,
		})
		diskCapacity.With(diskLabels).Set(float64(*disk.CapacityInMiB) * 1024 * 1024)
		diskUsed.With(diskLabels).Set(float64(*disk.AllocationInMiB) * 1024 * 1024)
	}
}

func (mu DefaultMetricsUpdater) updateMetricsFromServerInfos(serverInfos []collector.ServerInfo) {
	for _, serverInfo := range serverInfos {
		server := serverInfo.Server
		baseLabels := serverBaseLabels(server)

		// Update CPU and memory.
		cpuCores.With(baseLabels).Set(float64(*server.MaxCpuCount))
		memory.With(baseLabels).Set(float64(*server.ServerLiveInfo.MaxServerMemoryInMiB) * 1024 * 1024)

		// Update other server metrics (like uptime).
		serverStartTimeSeconds.With(baseLabels).Set(float64(*server.ServerLiveInfo.UptimeInSeconds))

		// Update server status.
		onlineStatus := SERVER_STATUS_ONLINE
		if *server.ServerLiveInfo.State == client.SHUTOFF {
			onlineStatus = SERVER_STATUS_OFFLINE
		}
		serverStatus.With(mergeLabels(baseLabels, prometheus.Labels{"status": onlineStatus.String()})).Set(float64(onlineStatus))

		// Update rescue system status.
		rescueStatus := RESCUE_SYSTEM_INACTIVE
		if *server.RescueSystemActive {
			rescueStatus = RESCUE_SYSTEM_ACTIVE
		}
		rescueActive.With(mergeLabels(baseLabels, prometheus.Labels{"status": rescueStatus.String()})).Set(float64(rescueStatus))

		// Update reboot recommendation status.
		rebootRecStatus := REBOOT_NOT_RECOMMENDED
		if !*server.ServerLiveInfo.LatestQemu {
			rebootRecStatus = REBOOT_RECOMMENDED
		}
		rebootRecommended.With(mergeLabels(baseLabels, prometheus.Labels{"status": rebootRecStatus.String()})).Set(float64(rebootRecStatus))

		// Update interface metrics.
		mu.updateInterfaceMetrics(server)

		// Update disk metrics.
		mu.updateDiskMetrics(server)
	}
}
