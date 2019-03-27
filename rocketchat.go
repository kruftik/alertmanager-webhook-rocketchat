package main

import (
	"fmt"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
)

type Config struct {
	Rocketchat  url.URL
	Credentials models.UserCredentials
}

func GetRocketChatAuthenticatedClient(configFile string) *realtime.Client {

	config := Config{}

	// Load the config from the file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		log.Fatalf("Error: %v", errYAML)
	}

	rtClient, errClient := realtime.NewClient(&config.Rocketchat, false)
	if errClient != nil {
		log.Printf("Error to get realtime client: %v", errClient)
		return nil
	}

	_, errUser := rtClient.Login(&config.Credentials)
	if errUser != nil {
		log.Printf("Error to login user: %v", errUser)
		return nil
	}

	return rtClient

}

// Function connects to RocketChat server, authenticate the user and send the notification
func SendNotification(rtClient *realtime.Client, data template.Data) {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {
		message := fmt.Sprintf("Alert: status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)
		log.Printf(message)

		_, errMessage := rtClient.SendMessage(channel, message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return
		}
	}
}
