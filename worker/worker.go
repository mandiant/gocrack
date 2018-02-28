package worker

import rpcclient "github.com/fireeye/gocrack/server/rpc/client"

var (
	// CompileTime is when this was compiled
	CompileTime string
	// CompileRev is the git revision hash (sha1)
	CompileRev string
)

// WorkerImpl describes the core functionality that the parent and child components must implement.
type WorkerImpl interface {
	Start() error
	Stop() error
}

// InitRPCChannel returns an RPCClient that communicates to the server for beacons, task payload retrieval, status updates, etc
func InitRPCChannel(cfg Config) (*rpcclient.RPCClient, error) {
	client := rpcclient.NewRPCClient(cfg.ServerConn.Address)
	if err := client.AddCredentials(
		[]byte(cfg.ServerConn.Certificate),
		[]byte(cfg.ServerConn.PrivateKey),
		[]byte(cfg.ServerConn.CACertificate),
	); err != nil {
		return nil, err
	}

	if cfg.ServerConn.ServerName != nil {
		client.OverrideServerName(*cfg.ServerConn.ServerName)
	}

	return client, nil
}
