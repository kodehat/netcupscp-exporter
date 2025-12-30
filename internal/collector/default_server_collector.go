package collector

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/kodehat/netcupscp-exporter/internal/authenticator"
	"github.com/kodehat/netcupscp-exporter/internal/client"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

type DefaultServerCollector struct {
	AuthData   *authenticator.AuthData
	httpClient http.Client
}

var _ ServerCollector = DefaultServerCollector{}

func NewDefaultServerCollector(authData *authenticator.AuthData) DefaultServerCollector {
	return DefaultServerCollector{
		AuthData:   authData,
		httpClient: http.Client{},
	}
}

func (c DefaultServerCollector) CollectServerData(context context.Context) ([]ServerInfo, error) {
	bearerAuth, err := securityprovider.NewSecurityProviderBearerToken(c.AuthData.AccessToken)
	if err != nil {
		slog.Error("error creating bearer auth", "error", err)
		return nil, err
	}
	// TODO: Generate API with url.
	respClient, err := client.NewClientWithResponses("https://servercontrolpanel.de/scp-core", client.WithHTTPClient(&c.httpClient), client.WithRequestEditorFn(bearerAuth.Intercept))
	if err != nil {
		slog.Error("error creating client", "error", err)
		return nil, err
	}
	resp, err := respClient.GetApiPingWithResponse(context)
	if err != nil {
		slog.Error("error pinging API", "error", err)
		return nil, err
	}
	slog.Debug("api ping status", "status", resp.Status())
	serverListMinimal, err := c.getServerListMinimal(context, respClient)
	if err != nil {
		slog.Error("error getting server list", "error", err)
		return nil, err
	}
	var servers = make([]ServerInfo, len(*serverListMinimal))
	for i, srv := range *serverListMinimal {
		server, err := c.getServer(context, *srv.Id, respClient)
		if err != nil {
			slog.Debug("error getting server", "serverId", *srv.Id, "error", err)
			return nil, err
		}
		servers[i] = ServerInfo{Server: server}
	}
	return servers, nil
}

func (c DefaultServerCollector) getServerListMinimal(context context.Context, respClient *client.ClientWithResponses) (*[]client.ServerListMinimal, error) {
	serversResp, err := respClient.GetApiV1ServersWithResponse(context, &client.GetApiV1ServersParams{})
	if err != nil {
		return nil, err
	}
	return serversResp.JSON200, nil
}

func (c DefaultServerCollector) getServer(context context.Context, serverId int32, respClient *client.ClientWithResponses) (*client.Server, error) {
	server, err := respClient.GetApiV1ServersServerIdWithResponse(context, serverId, &client.GetApiV1ServersServerIdParams{})
	if err != nil {
		return nil, err
	}
	return server.JSON200, nil
}
