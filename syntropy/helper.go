package syntropy

import (
	"context"
	"fmt"
	"github.com/SyntropyNet/syntropy-sdk-go/syntropy"
	"strings"
	"time"
)

func nullableStringToString(s syntropy.NullableString) string {
	val := s.Get()
	if val == nil {
		return ""
	}
	return *val
}

func nullableAgentStatusToString(s syntropy.NullableAgentStatus) string {
	val := s.Get()
	if val == nil {
		return ""
	}
	return val.AgentStatus
}

func convertAgentTagsToTfValue(in []syntropy.AgentTag) []Tag {
	var out []Tag
	for _, tag := range in {
		out = append(out, Tag{
			ID:   int64(tag.AgentTagId),
			Name: tag.AgentTagName,
		})
	}
	return out
}

func int64ArrayToInt32Array(arr []int64) []int32 {
	ret := make([]int32, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, int32(v))
	}
	return ret
}

func stringArrayToAgentTypeArray(arr []string) []syntropy.AgentType {
	ret := make([]syntropy.AgentType, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, syntropy.AgentType(v))
	}
	return ret
}

func stringArrayToAgentStatusArray(arr []string) []syntropy.AgentFilterAgentStatus {
	ret := make([]syntropy.AgentFilterAgentStatus, 0, len(arr))
	for _, v := range arr {
		ret = append(ret, syntropy.AgentFilterAgentStatus(v))
	}
	return ret
}

func tfValueToDateP(date string) (*time.Time, error) {
	layout := "2014-09-12T11:45:26.371Z"
	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func sumOfNaturalNumbers(n int) (sum int) {
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}

func getOneConnectionDetails(ctx context.Context, clt syntropy.ConnectionsApiService, connectionIDs int32) (*Connection, error) {
	connections, err := parseConnectionServices(clt.V1NetworkConnectionsServicesGet(ctx), []int32{connectionIDs})
	if err != nil {
		return nil, err
	}
	if len(connections) != 1 {
		return nil, fmt.Errorf("something went wrong. Expected 1 connection but got %d", len(connections))
	}
	return &connections[0], nil
}

func getMultipleConnectionDetails(ctx context.Context, clt syntropy.ConnectionsApiService, connectionIDs []int32) ([]Connection, error) {
	return parseConnectionServices(clt.V1NetworkConnectionsServicesGet(ctx), connectionIDs)
}

func parseConnectionServices(clt syntropy.ApiV1NetworkConnectionsServicesGetRequest, connectionIDs []int32) ([]Connection, error) {
	connS := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(connectionIDs)), ","), "[]")
	remote, _, err := clt.Filter(connS).Execute()
	if err != nil {
		return nil, fmt.Errorf("error while getting network connection service: %e", err)
	}

	var connections []Connection
	// Loop through all connections
	for _, connection := range remote.Data {
		var services []ConnectionServiceData
		// Loop through agents in that connection. One connection has 2 separate agents
		for _, agent := range []syntropy.V1ConnectionServiceAgent{connection.Agent1, connection.Agent2} {
			// Loop through agent services
			for _, service := range agent.AgentServices {
				// Loop through service subnets (id of subnet will be used to enable specific service/subnet)
				for _, subnet := range service.AgentServiceSubnets {
					// Check if subnet is enabled
					enabled := false
					for _, enabledSubnet := range connection.AgentConnectionSubnets {
						if enabledSubnet.AgentServiceSubnetId == subnet.AgentServiceSubnetId {
							enabled = enabledSubnet.AgentConnectionSubnetIsEnabled
							break
						}
					}
					services = append(services, ConnectionServiceData{
						ID:           int64(subnet.AgentServiceSubnetId),
						Name:         service.AgentServiceName,
						IP:           subnet.AgentServiceSubnetIp,
						Type:         string(service.AgentServiceType),
						Enabled:      enabled,
						AgentID:      int64(agent.AgentId),
						ConnectionId: int64(connection.AgentConnectionGroupId),
					})
				}
			}
		}
		connections = append(connections, Connection{
			Agent1ID:          connection.Agent1.AgentId,
			Agent2ID:          connection.Agent2.AgentId,
			ConnectionGroupID: connection.AgentConnectionGroupId,
			Services:          services,
		})
	}
	return connections, nil
}
