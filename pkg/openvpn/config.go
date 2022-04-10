package openvpn

import (
	"log"
	"os"
	"path"
	"text/template"
)

const configDir = "/var/lib/vpnmux/openvpn"

func init() {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Panicf("error creating openvpn config dir: %v", err)
	}
}

type Config struct {
	Dir          string
	Dev          string
	Proto        string
	Host         string
	Port         int
	MTU          int
	MTUExtra     int
	MSSFix       int
	Ping         int
	PingRestart  int
	RenegSec     int
	CredsFile    string
	Verb         int
	Cipher       string
	Auth         string
	KeyDirection int
	CACert       string
	TLSCert      string
}

type ConfigOptions struct {
	Host    string
	User    string
	Pass    string
	CACert  string
	TLSCert string
}

func NewConfig2(dir string, opts ConfigOptions) (*Config, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	credsFile, err := os.OpenFile(path.Join(dir, "creds"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return nil, err
	}
	defer credsFile.Close()

	credsTmpl, err := template.New("creds").Parse(credsTemplate)
	if err != nil {
		// TODO: panic?
		return nil, err
	}

	if err := credsTmpl.Execute(credsFile, struct {
		Username string
		Password string
	}{
		Username: opts.User,
		Password: opts.Pass,
	}); err != nil {
		return nil, err
	}

	configFile, err := os.OpenFile(path.Join(dir, "openvpn.conf"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	configTmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		// TODO: panic?
		return nil, err
	}

	c := &Config{
		Dir:          dir,
		Dev:          "tun",
		Proto:        "udp",
		Host:         opts.Host,
		Port:         1194,
		MTU:          1500,
		MTUExtra:     32,
		MSSFix:       1450,
		Ping:         15,
		PingRestart:  0,
		RenegSec:     0,
		CredsFile:    "creds",
		Verb:         3,
		Cipher:       "AES-256-CBC",
		Auth:         "SHA512",
		KeyDirection: 1,
		CACert:       opts.CACert,
		TLSCert:      opts.TLSCert,
	}

	if err := configTmpl.Execute(configFile, c); err != nil {
		return nil, err
	}

	return c, nil
}

// TODO: delete NewConfigFromName and NewConfig after deprecating v0 api
func NewConfigFromName(name string) (*Config, error) {
	dir := path.Join(configDir, name)
	// TODO: verify directory actually exists?
	return &Config{
		Dir: dir,
	}, nil
}

// Extremely limited in scope for now; just support configurable host
// with everything else hard-coded for a certain VPN provider
func NewConfig(name, host, user, pass string) (*Config, error) {
	dir := path.Join(configDir, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	credsFile, err := os.OpenFile(path.Join(dir, "creds"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return nil, err
	}
	defer credsFile.Close()

	credsTmpl, err := template.New("creds").Parse(credsTemplate)
	if err != nil {
		// TODO: panic?
		return nil, err
	}

	if err := credsTmpl.Execute(credsFile, struct {
		Username string
		Password string
	}{
		Username: user,
		Password: pass,
	}); err != nil {
		return nil, err
	}

	configFile, err := os.OpenFile(path.Join(dir, "openvpn.conf"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	configTmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		// TODO: panic?
		return nil, err
	}

	c := &Config{
		Dir:          dir,
		Dev:          "tun",
		Proto:        "udp",
		Host:         host,
		Port:         1194,
		MTU:          1500,
		MTUExtra:     32,
		MSSFix:       1450,
		Ping:         15,
		PingRestart:  0,
		RenegSec:     0,
		CredsFile:    "creds",
		Verb:         3,
		Cipher:       "AES-256-CBC",
		Auth:         "SHA512",
		KeyDirection: 1,
		CACert:       caCert,
		TLSCert:      openVPNKey,
	}

	if err := configTmpl.Execute(configFile, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Close() error {
	return os.RemoveAll(c.Dir)
}
