package collector

import (
	"context"
	"log/slog"
	"net/http"

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
	resp, err := respClient.GetApiPingWithResponse(ctx)
	if err != nil {
		slog.Error("error pinging API", "error", err)
		return nil, err
	}
	slog.Debug("api ping status", "status", resp.Status())
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

func (c DefaultServerCollector) getServerListMinimal(ctx context.Context, respClient *client.ClientWithResponses) (*[]client.ServerListMinimal, error) {
	serversResp, err := respClient.GetApiV1ServersWithResponse(ctx, &client.GetApiV1ServersParams{})
	if err != nil {
		return nil, err
	}
	return serversResp.JSON200, nil
}

func (c DefaultServerCollector) getServer(ctx context.Context, serverId int32, respClient *client.ClientWithResponses) (*client.Server, error) {
	server, err := respClient.GetApiV1ServersServerIdWithResponse(ctx, serverId, &client.GetApiV1ServersServerIdParams{})
	if err != nil {
		return nil, err
	}
	return server.JSON200, nil
}
