package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/client"
	"github.com/kodehat/netcupscp-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	SERVER_ONLINE_STR  string  = "online"
	SERVER_OFFLINE_STR string  = "offline"
	SERVER_ONLINE_NUM  float64 = 1
	SERVER_OFFLINE_NUM float64 = 0

	RESCUE_SYSTEM_ACTIVE_NUM   float64 = 1
	RESCUE_SYSTEM_INACTIVE_NUM float64 = 0
	REBOOT_RECOMMENDED_NUM     float64 = 1
	REBOOT_NOT_RECOMMENDED_NUM float64 = 0
	DISK_OPTIMIZATION_YES_NUM  float64 = 1
	DISK_OPTIMIZATION_NO_NUM   float64 = 0
)

type MetricsUpdater struct {
	authData  *authenticator.AuthData
	collector collector.ServerCollector
}

func NewMetricsUpdater(authData *authenticator.AuthData) *MetricsUpdater {
	return &MetricsUpdater{
		authData:  authData,
		collector: collector.NewDefaultServerCollector(authData),
	}
}

func (mu MetricsUpdater) UpdateMetricsPeriodically(context context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	metricsUpdateFunc := func() {
		err := mu.UpdateMetrics(context)
		if err != nil {
			fmt.Printf("Error updating metrics: %v\n", err)
		} else {
			fmt.Println("Metrics updated successfully")
		}
	}
	metricsUpdateFunc() // Run once immediately.
	for {
		select {
		case <-ticker.C:
			metricsUpdateFunc()
		case <-context.Done():
			fmt.Println("Stopping metrics updater")
			return
		}
	}
}

func (mu MetricsUpdater) UpdateMetrics(context context.Context) error {
	serverInfos, err := mu.collector.CollectServerData(context)
	if err != nil {
		return err
	}
	mu.updateMetricsFromServerInfos(serverInfos)
	return nil
}

func (mu MetricsUpdater) updateInterfaceMetrics(server *client.Server) {
	for _, iface := range *server.ServerLiveInfo.Interfaces {
		monthlyTrafficIn.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.RxMonthlyInMiB) * 1024 * 1024)
		monthlyTrafficOut.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.TxMonthlyInMiB) * 1024 * 1024)
		monthlyTrafficTotal.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.TxMonthlyInMiB+*iface.RxMonthlyInMiB) * 1024 * 1024)
		var throttled float64
		throttled = 0
		if *iface.TrafficThrottled {
			throttled = 1
		}
		ifaceThrottled.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "mac": *iface.Mac}).Set(throttled)
		for _, ip := range *iface.Ipv4Addresses {
			serverIpInfo.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "mac": *iface.Mac, "ip": ip, "type": "ipv4"}).Set(1)
		}
		for _, ip := range *iface.Ipv6NetworkPrefixes {
			serverIpInfo.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "mac": *iface.Mac, "ip": ip, "type": "ipv6"}).Set(1)
		}
	}
}

func (mu MetricsUpdater) updateDiskMetrics(server *client.Server) {
	var diskOpt float64
	diskOpt = DISK_OPTIMIZATION_YES_NUM
	if *server.ServerLiveInfo.RequiredStorageOptimization == client.NO {
		diskOpt = DISK_OPTIMIZATION_NO_NUM
	}
	for _, disk := range *server.ServerLiveInfo.Disks {
		diskCapacity.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "driver": *disk.Driver, "name": *disk.Dev}).Set(float64(*disk.CapacityInMiB) * 1024 * 1024)
		diskUsed.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "driver": *disk.Driver, "name": *disk.Dev}).Set(float64(*disk.AllocationInMiB) * 1024 * 1024)
	}
	diskOptimization.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(diskOpt)
}

func (mu MetricsUpdater) updateMetricsFromServerInfos(serverInfos []collector.ServerInfo) {
	for _, serverInfo := range serverInfos {
		server := serverInfo.Server
		cpuCores.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(float64(*server.ServerLiveInfo.CpuMaxCount))
		memory.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(float64(*server.ServerLiveInfo.MaxServerMemoryInMiB) * 1024 * 1024)
		mu.updateInterfaceMetrics(server)
		serverStartTimeSeconds.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(float64(*server.ServerLiveInfo.UptimeInSeconds))
		var online float64
		online = SERVER_ONLINE_NUM
		status := SERVER_ONLINE_STR
		if *server.ServerLiveInfo.State == client.SHUTOFF {
			online = SERVER_OFFLINE_NUM
			status = SERVER_OFFLINE_STR
		}
		serverStatus.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname, "status": status}).Set(online)
		var rescue float64
		rescue = RESCUE_SYSTEM_INACTIVE_NUM
		if *server.RescueSystemActive {
			rescue = RESCUE_SYSTEM_ACTIVE_NUM
		}
		rescueActive.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(rescue)
		var rebootRec float64
		rebootRec = REBOOT_NOT_RECOMMENDED_NUM
		if !*server.ServerLiveInfo.LatestQemu {
			rebootRec = REBOOT_RECOMMENDED_NUM
		}
		rebootRecommended.With(prometheus.Labels{"vserver": *server.Name, "nickname": *server.Nickname}).Set(rebootRec)
		mu.updateDiskMetrics(server)
	}
}
