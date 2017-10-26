package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type IPool interface {
	Get() interface{}
	Put(x interface{})
}

type gzReader struct {
	sock io.ReadCloser
	gz   *gzip.Reader
	p    IPool
}

func (s *gzReader) Close() error {
	defer s.p.Put(s.gz)

	if err := s.sock.Close(); err != nil {
		return err
	}

	if err := s.gz.Close(); err != nil {
		return err
	}
	return nil
}

func (s *gzReader) Read(p []byte) (int, error) {
	return s.gz.Read(p)
}

type RPCClient struct {
	u  url.URL
	c  *http.Client
	tc *tls.Config
	p  sync.Pool // a pool of *gzip.Readers for decompression
}

type FileResponse struct {
	File io.ReadCloser
	Hash string
}

func NewRPCClient(serverAddress string) *RPCClient {
	return &RPCClient{
		u: url.URL{
			Host:   serverAddress,
			Scheme: "https",
		},
		c: &http.Client{},
		p: sync.Pool{
			New: func() interface{} {
				return new(gzip.Reader)
			},
		},
	}
}

func (s *RPCClient) OverrideServerName(serverName string) error {
	if s.tc == nil {
		return errors.New("cannot override server name when credentials have not been added to the client")
	}

	s.tc.ServerName = serverName
	return nil
}

func (s *RPCClient) AddCredentials(certificate, privateKey, caCertificate []byte) error {
	cert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if caCertificate != nil {
		certp := x509.NewCertPool()
		if ok := certp.AppendCertsFromPEM(caCertificate); !ok {
			return errors.New("failed to build cert pool with ca certificate")
		}
		tlsConfig.RootCAs = certp
	}

	tlsConfig.BuildNameToCertificate()
	s.c.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	s.tc = tlsConfig

	return nil
}

func (s *RPCClient) performJSONCall(method, path string, input interface{}, output interface{}) error {
	var buf io.ReadWriter

	if input != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(&input); err != nil {
			return err
		}
	}

	s.u.Path = path
	req, err := http.NewRequest(method, s.u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	res, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Nothing else to do!
	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
		gz := s.p.Get().(*gzip.Reader)
		defer s.p.Put(gz)
		if err := gz.Reset(res.Body); err != nil {
			return err
		}
		defer gz.Close()
		res.Body = gz
	}

	if strings.Contains(strings.ToLower(res.Header.Get("Content-Type")), "application/json") {
		return json.NewDecoder(res.Body).Decode(&output)
	}

	wr, ok := output.(io.Writer)
	if !ok {
		return errors.New("output must be a io.Writer")
	}

	_, err = io.Copy(wr, res.Body)
	return err
}

func (s *RPCClient) performFileCall(method, path string, input interface{}) (*FileResponse, error) {
	var buf io.ReadWriter

	if input != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(&input); err != nil {
			return nil, err
		}
	}

	s.u.Path = path
	req, err := http.NewRequest(method, s.u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	res, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}

	if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
		gz := s.p.Get().(*gzip.Reader)
		if err := gz.Reset(res.Body); err != nil {
			return nil, err
		}
		res.Body = &gzReader{sock: res.Body, gz: gz, p: &s.p}
	}

	return &FileResponse{
		File: res.Body,
		Hash: res.Header.Get("X-FileHash-SHA1"),
	}, nil
}
