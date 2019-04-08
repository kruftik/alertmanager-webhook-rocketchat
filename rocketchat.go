package main

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/RocketChat/Rocket.Chat.Go.SDK/realtime"
	"github.com/prometheus/alertmanager/template"
)

type Config struct {
	Endpoint  url.URL
	Credentials models.UserCredentials
}

type RocketChatClient interface {
	GetChannelId(name string) (string, error)
	SendMessage(channel *models.Channel, text string) (*models.Message, error)
}

func GetRocketChatAuthenticatedClient(config Config) *realtime.Client {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
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

// SendNotification connects to RocketChat server, authenticate the user and send the notification
func SendNotification(rtClient RocketChatClient, data template.Data) {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {
		message := formatMessage(alert)
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

	var keys []string
	for k := range alert.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		msgContent += fmt.Sprintf("**%s**: %s\n", k, alert.Labels[k])
	}

	return msgContent
}
