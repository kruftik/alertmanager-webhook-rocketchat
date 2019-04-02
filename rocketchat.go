package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
)

type Config struct {
	Rocketchat  url.URL
	Credentials models.UserCredentials
}

type RocketChatClient interface {
	GetChannelId(name string) (string, error)
	SendMessage(channel *models.Channel, text string) (*models.Message, error)
}

func GetRocketChatAuthenticatedClient(config Config) *realtime.Client {

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
func SendNotification(rtClient RocketChatClient, data template.Data) {

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

func formatMessage(alert template.Alert) string {

	msgContent := fmt.Sprintf("**%s: %s**\n", strings.Title(alert.Status), alert.Annotations["summary"])
	msgContent += fmt.Sprintf("**description**: %s\n", alert.Annotations["description"])

	for k, v := range alert.Labels {
		msgContent += fmt.Sprintf("**%s**: %s\n", k, v)
	}

	return msgContent
}
