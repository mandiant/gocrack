package client

import "github.com/fireeye/gocrack/server/rpc"

// Beacon sends our current status to the server
func (s *RPCClient) Beacon(beaconRequest rpc.BeaconRequest) (*rpc.BeaconResponse, error) {
	var resp rpc.BeaconResponse

	if err := s.performJSONCall("POST", "/rpc/v1/beacon", beaconRequest, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
