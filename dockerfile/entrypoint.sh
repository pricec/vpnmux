#!/bin/sh

iptables -t nat -A POSTROUTING -o tun0 -j MASQUERADE
iptables -t filter -A FORWARD -i eth0 -o !tun0 -j DROP

GATEWAY_IP=$(ip route show default | sed -E 's/.*via ([0-9.]+) dev.*/\1/')
ip route add ${LOCAL_SUBNET_CIDR} via ${GATEWAY_IP}

openvpn $@
