FROM alpine:3.15.4

RUN apk add openvpn
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
