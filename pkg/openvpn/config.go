package openvpn

import (
	"os"
	"path"
	"text/template"
)

// TODO: make this configutable
const configDir = "/var/lib/vpnmux/openvpn"

type Config struct {
	ID           string
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

// TODO: recover settings?
func NewConfigFromID(id string) (*Config, error) {
	dir := path.Join(configDir, id)
	creds := path.Join(dir, "creds")
	ovpn := path.Join(dir, "openvpn.conf")

	if _, err := os.Stat(creds); err != nil {
		return nil, err
	}

	if _, err := os.Stat(ovpn); err != nil {
		return nil, err
	}

	return &Config{
		ID:  id,
		Dir: dir,
	}, nil
}

func NewConfig(id string, opts ConfigOptions) (*Config, error) {
	c := &Config{
		ID:           id,
		Dir:          "",
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

	dir := path.Join(configDir, id)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	c.Dir = dir

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

	if err := configTmpl.Execute(configFile, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Close() error {
	return os.RemoveAll(c.Dir)
}
