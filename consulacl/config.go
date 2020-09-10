package consulacl

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/logging"
)

type Config struct {
	// Destination
	Token string `mapstructure:"token"`
	// // Auth
	Address string `mapstructure:"address"`
	// TLS
	Scheme        string `mapstructure:"scheme"`
	CAFile        string `mapstructure:"ca_file"`
	CertFile      string `mapstructure:"cert_file"`
	KeyFile       string `mapstructure:"key_file"`
	TlsSkipVerify bool   `mapstructure:"tls_skip_verify"`
}

func (c *Config) Client() (*consul.Client, error) {
	config := consul.DefaultConfig()
	if c.Address != "" {
		config.Address = c.Address
	}
	if c.Scheme != "" {
		config.Scheme = c.Scheme
	}

	tlsConfig := consul.TLSConfig{}
	tlsConfig.CAFile = c.CAFile
	tlsConfig.CertFile = c.CertFile
	tlsConfig.KeyFile = c.KeyFile
	tlsConfig.InsecureSkipVerify = c.TlsSkipVerify

	var err error
	config.HttpClient, err = consul.NewHttpClient(config.Transport, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: '%s'", err)
	}

	if c.Token != "" {
		config.Token = c.Token
	}

	if logging.IsDebugOrHigher() {
		config.HttpClient.Transport = logging.NewTransport(
			"consulacl",
			config.HttpClient.Transport,
		)
	}

	client, err := consul.NewClient(config)

	if err != nil {
		return nil, err
	}
	return client, nil
}
