package openvpn

var credsTemplate = `{{.Username}}
{{.Password}}
`

var configTemplate = `
client
dev {{.Dev}}
proto {{.Proto}}
remote {{.Host}} {{.Port}}
resolv-retry infinite
remote-random
nobind
tun-mtu {{.MTU}}
tun-mtu-extra {{.MTUExtra}}
mssfix {{.MSSFix}}
persist-key
persist-tun
ping {{.Ping}}
ping-restart {{.PingRestart}}
ping-timer-rem
reneg-sec {{.RenegSec}}

remote-cert-tls server

auth-user-pass {{.CredsFile}}
verb {{.Verb}}
pull
fast-io
cipher {{.Cipher}}

auth {{.Auth}}

<ca>
{{.CACert}}
</ca>
key-direction {{.KeyDirection}}
<tls-auth>
{{.TLSCert}}
</tls-auth>
`
