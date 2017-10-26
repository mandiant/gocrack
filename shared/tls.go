package shared

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

// GetTLSConfig is a helper function to return a TLS configuration that can be used in TLS servers
func GetTLSConfig(certificatePem, privateKeyPem string, caCertPem *string) (*tls.Config, error) {
	cert, err := tls.X509KeyPair([]byte(certificatePem), []byte(privateKeyPem))
	if err != nil {
		return nil, err
	}

	tcfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if caCertPem != nil && *caCertPem != "" {
		certp := x509.NewCertPool()
		if ok := certp.AppendCertsFromPEM([]byte(*caCertPem)); !ok {
			return nil, errors.New("failed to build cert pool with ca certificate")
		}
		tcfg.RootCAs = certp
	}

	return tcfg, nil
}
