FROM golang:1.19-alpine as builder
WORKDIR /alertmanager-webhook-rocketchat
RUN apk add --no-cache git
COPY . .
RUN go build -o ./app ./

FROM alpine:3 AS app
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"
COPY --from=builder /alertmanager-webhook-rocketchat/app /bin/alertmanager-webhook-rocketchat

EXPOSE      9876

WORKDIR /
ENTRYPOINT  [ "/bin/alertmanager-webhook-rocketchat" ]
