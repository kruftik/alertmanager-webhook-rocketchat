# alertmanager-webhook-rocketchat
Prometheus AlertManager webhook receiver to Rocket.Chat

https://travis-ci.org/FXinnovation/alertmanager-webhook-rocketchat.svg?branch=master

The project takes 2 optional parameters to be configured :
- config.file to specify RocketChat configuration. Cf config/rocketchat_example.yml (default : config/rocketchat.yml)
- listen.address to specify the listening port (default : 9876)