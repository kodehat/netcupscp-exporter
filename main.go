package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/client"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type AuthInfo struct {
	AccessToken  string
	RefreshToken string
}

type ServerInfo struct {
	*client.Server
}

var (
	netcupAuth  AuthInfo
	host        string
	port        string
	serverInfos []ServerInfo
	cpuCores    = prometheus.NewGaugeVec(
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

func updateData() {
	bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(netcupAuth.AccessToken)
	hc := http.Client{}
	c, err := client.NewClientWithResponses("https://servercontrolpanel.de/scp-core", client.WithHTTPClient(&hc), client.WithRequestEditorFn(bearerAuth.Intercept))
	if err != nil {
		log.Panicf("Error creating client: %v", err)
	}
	resp, err := c.GetApiPingWithResponse(context.TODO())
	if err != nil {
		log.Panicf("Error pinging API: %v", err)
	}
	log.Printf("API Ping Status: %s\n", resp.Status())
	serversResp, err := c.GetApiV1ServersWithResponse(context.TODO(), &client.GetApiV1ServersParams{})
	if err != nil {
		log.Panicf("Error getting servers: %v", err)
	}
	var servers = make([]ServerInfo, len(*serversResp.JSON200))
	for i, srv := range *serversResp.JSON200 {
		id := srv.Id
		server, err := c.GetApiV1ServersServerIdWithResponse(context.TODO(), *id, &client.GetApiV1ServersServerIdParams{})
		if err != nil {
			log.Panicf("Error getting server %d: %v", id, err)
		}
		servers[i] = ServerInfo{Server: server.JSON200}
	}
	serverInfos = servers
}

func updateMetrics() {
	for {
		authenticate()
		updateData()

		for _, server := range serverInfos {
			cpuCores.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(float64(*server.Server.ServerLiveInfo.CpuMaxCount))
			memory.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(float64(*server.Server.ServerLiveInfo.MaxServerMemoryInMiB) * 1024 * 1024)
			for _, iface := range *server.Server.ServerLiveInfo.Interfaces {
				monthlyTrafficIn.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.RxMonthlyInMiB) * 1024 * 1024)
				monthlyTrafficOut.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.TxMonthlyInMiB) * 1024 * 1024)
				monthlyTrafficTotal.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "month": fmt.Sprintf("%02d", time.Now().Month()), "year": fmt.Sprintf("%d", time.Now().Year()), "mac": *iface.Mac}).Set(float64(*iface.TxMonthlyInMiB+*iface.RxMonthlyInMiB) * 1024 * 1024)
				var throttled float64
				throttled = 0
				if *iface.TrafficThrottled {
					throttled = 1
				}
				ifaceThrottled.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "mac": *iface.Mac}).Set(throttled)
				for _, ip := range *iface.Ipv4Addresses {
					serverIpInfo.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "mac": *iface.Mac, "ip": ip, "type": "ipv4"}).Set(1)
				}
				for _, ip := range *iface.Ipv6NetworkPrefixes {
					serverIpInfo.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "mac": *iface.Mac, "ip": ip, "type": "ipv6"}).Set(1)
				}
			}
			serverStartTimeSeconds.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(float64(*server.Server.ServerLiveInfo.UptimeInSeconds))
			var online float64
			online = 1
			status := "online"
			if *server.Server.ServerLiveInfo.State == client.SHUTOFF {
				online = 0
				status = "offline"
			}
			serverStatus.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "status": status}).Set(online)
			var rescue float64
			rescue = 0
			if *server.Server.RescueSystemActive {
				rescue = 1
			}
			rescueActive.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(rescue)
			var rebootRec float64
			rebootRec = 0
			if !*server.Server.ServerLiveInfo.LatestQemu {
				rebootRec = 1
			}
			rebootRecommended.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(rebootRec)
			var diskOpt float64
			diskOpt = 1
			if *server.Server.ServerLiveInfo.RequiredStorageOptimization == client.NO {
				diskOpt = 0
			}
			for _, disk := range *server.Server.ServerLiveInfo.Disks {
				diskCapacity.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "driver": *disk.Driver, "name": *disk.Dev}).Set(float64(*disk.CapacityInMiB) * 1024 * 1024)
				diskUsed.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname, "driver": *disk.Driver, "name": *disk.Dev}).Set(float64(*disk.AllocationInMiB) * 1024 * 1024)
			}
			diskOptimization.With(prometheus.Labels{"vserver": *server.Server.Name, "nickname": *server.Nickname}).Set(diskOpt)
		}

		// Try to avoid rate limiting
		// Limit is 200req / 1h
		time.Sleep(60 * time.Second)
	}
}

func authenticate() {
	issuer := "https://www.servercontrolpanel.de/realms/scp"
	clientId := "scp"
	scopes := strings.Split("offline_access openid", " ")

	provider, err := rp.NewRelyingPartyOIDC(context.TODO(), issuer, clientId, "", "", scopes)
	if err != nil {
		log.Fatalf("Error creating OIDC provider: %v", err)
	}
	if netcupAuth.RefreshToken == "" {
		resp, err := rp.DeviceAuthorization(context.TODO(), scopes, provider, nil)
		if err != nil {
			log.Fatalf("Error during device authorization: %v", err)
		}
		fmt.Printf("\nPlease browse to %s and enter code %s\n", resp.VerificationURI, resp.UserCode)
		token, err := rp.DeviceAccessToken(context.TODO(), resp.DeviceCode, time.Duration(resp.Interval)*time.Second, provider)
		if err != nil {
			log.Fatalf("Error getting access token: %v", err)
		}
		fmt.Printf("Access Token: %s\n", token.AccessToken)
		fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
		netcupAuth = AuthInfo{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
		}
	}
	token, err := rp.RefreshTokens[*oidc.IDTokenClaims](context.TODO(), provider, netcupAuth.RefreshToken, "", "")
	if err != nil {
		log.Fatalf("Error refreshing token: %v", err)
	}
	netcupAuth = AuthInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}

func main() {
	refreshToken := os.Getenv("NETCUP_REFRESH_TOKEN")
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	if host == "" {
		log.Println("No host provided, using default (bind to all interfaces)")
	}
	if port == "" {
		log.Println("No port provided, using default (9509)")
		port = "9509"
	}
	netcupAuth = AuthInfo{
		AccessToken:  "",
		RefreshToken: refreshToken,
	}

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

	go updateMetrics()
	listenAddr := net.JoinHostPort(host, port)
	fmt.Printf("Listening on %q\n", listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}
