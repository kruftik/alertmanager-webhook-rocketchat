FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Thomas Lacher <thomas.lacher@fxinnovation.com>

COPY .  /bin/alertmanager-webhook-rocketchat

EXPOSE      9876
ENTRYPOINT [ "/bin/alertmanager-webhook-rocketchat" ]