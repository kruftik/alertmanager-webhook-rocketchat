FROM golang:1.19-alpine as builder
WORKDIR /alertmanager-webhook-rocketchat
COPY . .
RUN go build -o ./alertmanager-webhook-rocketchat ./

FROM alpine:3 AS app
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"
COPY --from=builder /alertmanager-webhook-rocketchat /bin/alertmanager-webhook-rocketchat

EXPOSE      9876

WORKDIR /
ENTRYPOINT  [ "/bin/alertmanager-webhook-rocketchat" ]
