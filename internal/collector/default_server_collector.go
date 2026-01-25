package collector

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/client"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

const scpBaseUrl = "https://servercontrolpanel.de/scp-core"

type DefaultServerCollector struct {
	authenticator authenticator.Authenticator
	httpClient    http.Client
}

var _ ServerCollector = DefaultServerCollector{}

func NewDefaultServerCollector(authenticator authenticator.Authenticator) DefaultServerCollector {
	return DefaultServerCollector{
		authenticator: authenticator,
		httpClient:    http.Client{},
	}
}

func (c DefaultServerCollector) CollectServerData(ctx context.Context) ([]ServerInfo, error) {
	// Prepare authenticated client.
	bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(c.authenticator.GetAuthData().AccessToken)
	if err != nil {
		slog.Error("error creating bearer auth", "error", err)
		return nil, err
	}
	respClient, err := client.NewClientWithResponses(scpBaseUrl, client.WithHTTPClient(&c.httpClient), client.WithRequestEditorFn(bearerAuth.Intercept))
	if err != nil {
		slog.Error("error creating client", "error", err)
		return nil, err
	}

	// Check API availability.
	pingStatus, err := c.pingApi(ctx, respClient)
	if err != nil {
		slog.Error("error pinging API", "error", err)
		return nil, err
	}
	slog.Debug("api ping status", "status", pingStatus)

	// Get maintenance info and check if maintenance is ongoing.
	// HINT: Disabled as Netcup does not return a list of maintenance information, but only a single object.
	// Despite the API docs specifying an array of maintenance objects.
	/* 	maintenanceInfo, err := c.getMaintenanceInfo(ctx, respClient)
	   	if err != nil {
	   		slog.Error("error getting maintenance info", "error", err)
	   		return nil, err
	   	}
	   	if maintenanceOngoing, maintenanceInfo := isMaintenanceOngoing(maintenanceInfo, time.Now()); maintenanceOngoing {
	   		slog.Warn("maintenance is currently ongoing", "start_at", maintenanceInfo.StartAt, "finish_at", maintenanceInfo.FinishAt)
	   		return nil, errors.New("maintenance is currently ongoing")
	   	}
	   	slog.Debug("no ongoing maintenance detected") */

	serverListMinimal, err := c.getServerListMinimal(ctx, respClient)
	if err != nil {
		slog.Error("error getting server list", "error", err)
		return nil, err
	}
	if serverListMinimal == nil {
		slog.Warn("no servers found")
		return []ServerInfo{}, nil
	}
	var servers = make([]ServerInfo, len(*serverListMinimal))
	for i, srv := range *serverListMinimal {
		server, err := c.getServer(ctx, *srv.Id, respClient)
		if err != nil {
			slog.Debug("error getting server", "serverId", *srv.Id, "error", err)
			return nil, err
		}
		servers[i] = ServerInfo{Server: server}
	}
	return servers, nil
}

func isMaintenanceOngoing(maintenanceList *[]client.Maintenance, compareVal time.Time) (bool, *client.Maintenance) {
	if maintenanceList == nil {
		return false, nil
	}
	for _, maintenance := range *maintenanceList {
		if TimeBetween(compareVal, maintenance.StartAt, maintenance.FinishAt) {
			return true, &maintenance
		}
	}
	return false, nil
}

func (c DefaultServerCollector) pingApi(ctx context.Context, respClient *client.ClientWithResponses) (string, error) {
	pingResp, err := respClient.GetApiPingWithResponse(ctx)
	if err != nil {
		return "", err
	} else if pingResp.StatusCode() != http.StatusOK {
		return "", errors.New("unexpected status code when pinging API: " + pingResp.Status())
	}
	return pingResp.Status(), nil
}

func (c DefaultServerCollector) getMaintenanceInfo(ctx context.Context, respClient *client.ClientWithResponses) (*[]client.Maintenance, error) {
	maintenanceResp, err := respClient.GetApiV1MaintenanceWithResponse(ctx)
	if err != nil {
		return nil, err
	} else if maintenanceResp.StatusCode() != http.StatusOK {
		return nil, errors.New("unexpected status code when getting maintenance info: " + maintenanceResp.Status())
	}
	return maintenanceResp.JSON200, nil
}

func (c DefaultServerCollector) getServerListMinimal(ctx context.Context, respClient *client.ClientWithResponses) (*[]client.ServerListMinimal, error) {
	serversResp, err := respClient.GetApiV1ServersWithResponse(ctx, &client.GetApiV1ServersParams{})
	if err != nil {
		return nil, err
	} else if serversResp.StatusCode() != http.StatusOK {
		return nil, errors.New("unexpected status code when getting server list: " + serversResp.Status())
	}
	return serversResp.JSON200, nil
}

func (c DefaultServerCollector) getServer(ctx context.Context, serverId int32, respClient *client.ClientWithResponses) (*client.Server, error) {
	server, err := respClient.GetApiV1ServersServerIdWithResponse(ctx, serverId, &client.GetApiV1ServersServerIdParams{})
	if err != nil {
		return nil, err
	} else if server.StatusCode() != http.StatusOK {
		return nil, errors.New("unexpected status code when getting server: " + server.Status())
	}
	return server.JSON200, nil
}
