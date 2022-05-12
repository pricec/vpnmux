# vpnmux - VPN Multiplexer
The purpose of this application is to allow easy and dynamic control of a
series of VPN connections on a gateway device, and the assignment of specific
hosts to specific VPN connections. When a host is assigned to a certain VPN
connection, all traffic for that host is routed across the connection; if
the connection fails, the host will lose connectivity.

# Installation & Configuration
```bash
VERSION=0.2.1-beta1

curl -Lo /usr/local/bin/vpnmux-${VERSION} https://github.com/pricec/vpnmux/releases/download/${VERSION}/vpnmux-${VERSION}

cat > /etc/systemd/system/vpnmux.service <<EOF
[Unit]
Description=VPN multiplexer
Wants=network.target
After=networking.service
Before=network.target network-online.target

[Service]
EnvironmentFile=/var/lib/vpnmux/config.env
ExecStart=/usr/local/bin/vpnmux-${VERSION}
Restart=always

[Install]
WantedBy=multi-user.target
WantedBy=network-online.target
EOF

cat > /var/lib/vpnmux/config.env <<EOF
# (optional) Path to sqlite database file; used for storing resources
# (default=/var/lib/vpnmux/v1.db)
VPNMUX_DB_PATH=/var/lib/vpnmux/v1.db
# (optional) Docker container image to use for VPN containers
# (default=pricec/openvpn-client:latest)
VPNMUX_IMAGE=pricec/openvpn-client:latest
# Local subnet CIDR; needed to route traffic from the VPN back
# to clients on the local network.
VPNMUX_SUBNET_CIDR=192.168.1.0/24
# (optional) HTTP server shutdown timeout (default=10s)
VPNMUX_SHUTDOWN_TIMEOUT=10s
# (optional) HTTP server listen port (default=8080)
VPNMUX_LISTEN_PORT=8080
# LAN and WAN interfaces on gateway host. These interface names are used
# to create iptables rules preventing forwarding of packets from the
# LAN interface to the WAN interface for each client configured in vpnmux.
# This is a safety feature to ensure configured clients don't accidentally
# expose themselves when updating the vpnmux config or when a VPN is down.
VPNMUX_LAN_INTERFACE=lan0
VPNMUX_WAN_INTERFACE=wan0
EOF

systemctl daemon-reload
systemctl enable vpnmux.service
systemctl restart vpnmux.service
```

# API
## v1
The API is a standard REST API employing JSON, and works as you might expect.

### Credentials
The `Credential` resource is meant to store usernames, passwords, and
cryptographic keys for use in a `Config` resource. Note that the v1 `Config`
resource is highly prescriptive in what credentials it expects to be specified
in each instance.

The `Credential` resource has the following schema.
```json
{
    "id": "<string>",
    "name": "<string>",
    "value": "<string>"
}
```

The following endpoints are available.
* `GET /v1/credential` - returns a list of credentials. The `value` field
  is omitted.
* `GET /v1/credential/{id}` - returns the specified credential, or 404 if no
  such credential exists. The `value` field is present.
* `POST /v1/credential` - expects a `Credential` resource in the body; creates
  the resource in the server.
* `DELETE /v1/credential/{id}` - deletes the specified credential, or 404
  if no such credential exists.

### Configs
The `Config` resource represents an OpenVPN configuration. These are meant to
be generic, but your mileage may vary. Many options currently take on default
values and are not exposed, and a certain type of configuration is expected.
Namely, it is expected that username/password authentication is used, a CA
certificate is specified and a TLS auth certificate is provided. Each of these
must correspond to a `Credential` resource. The only other configurable item
is the remote host. For details, see `package openvpn` (`pkg/openvpn`).

The `Config` resource has the following schema.
```json
{
    "id": "<string>",
    "name": "<string>",
    "host": "<string>",
    "user_cred": "<Credential ID>",
    "pass_cred": "<Credential ID>",
    "ca_cred": "<Credential ID>",
    "ovpn_cred": "<Credential ID>"
}
```

The following endpoints are available.
* `GET /config` - returns a list of configs, just containing the `id` and
  `name` fields.
* `GET /config/{id}` - returns the specified config, or 404 if no such config
  exists. All fields are populated.
* `POST /config` - expects a `Config` resource in the body; creates the
  resource in the server.
* `PATCH /config/{id}` - expects a `Config` resource in the body; updates the
  config in the path accordingly. The `id` field in the body is ignored.
* `DELETE /config/{id}` - deletes the specified config, or 404 if no such
  config exists.

### Networks
A `Network` is a resource representing an OpenVPN connection. It is so named
because it also corresponds to a docker (bridge) network hosting the container;
one docker network and one container is created for each such resource. We
also create a routing table for each `Network` and assign the default route
via the OpenVPN container address.

The `Network` resource has the following schema.
```json
{
    "id": "<string>",
    "name": "<string>",
    "config_id": "<Config ID>"
}
```

The following endpoints are available.
* `GET /v1/network` - returns a list of networks containing all fields.
* `GET /v1/network/{id}` - returns the specified network, or 404 if no such
  network exists.
* `POST /v1/network` - expects a `Network` resource in the body; creates the
  corresponding docker network, container, and routing table, and creates the
  resource in the server.
* `PATCH /v1/network/{id}` - expects a `Network` resource in the body; updates
  the network in the path accordingly. The `id` field in the body is ignored.
* `DELETE /v1/network/{id}` - deletes the specified network, or 404 if no such
  network exists.

### Clients
A `Client` resource represents a host on the network. When creating a `Client`,
the gateway (i.e. server host) will set up rules to ensure the corresponding
host's packets aren't routed onto the network, except via an OpenVPN client.

The `Client` resource has the following schema.
```json
{
    "id": "<string>",
    "name": "<string>",
    "address": "<string>"
}
```

The following endpoints are available.
* `GET /v1/client` - returns a list of clients containing all fields.
* `GET /v1/client/{id}` - returns the specified client, or 404 if no such
  client exists.
* `POST /v1/client` - expects a `Client` resource in the body; creates the
  corresponding rules to prevent forwarding of the host's packets directly
  onto the network.
* `PATCH /v1/client/{id}` - expects a `Client` resource in the body; updates
  the client in the path accordingly. The `id` field in the body is ignored.
* `DELETE /v1/client/{id}` - deletes the specified client, or 404 if no such
  client exists.

### Client Networks
It is possible to assign a `Client` to a `Network`, and when you do this, all
traffic from the corresponding host is routed via the corresponding OpenVPN
connection.

The Client-Network association has the following schema.
```json
{
    "client_id": "<Client ID>",
    "network_id": "<Network ID>"
}
```

The following endpoints are available.
* `GET /v1/client/{id}/network` - returns the Client-Network association for
  the given client, if one exists.
* `DELETE /v1/client/{id}/network` - unassigns the given client from its
  network.
* `POST /v1/client/{id}/network/{id}` - assigns the given client to the
  given network.