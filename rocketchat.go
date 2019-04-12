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
	Endpoint    url.URL
	Credentials models.UserCredentials
}

type RocketChatClient interface {
	Login(credentials *models.UserCredentials) (*models.User, error)
	GetChannelId(name string) (string, error)
	SendMessage(channel *models.Channel, text string) (*models.Message, error)
}

func GetRocketChatClient(config Config) (*realtime.Client, error) {

	rtClient, errClient := realtime.NewClient(&config.Endpoint, false)
	if errClient != nil {
		return nil, errClient
	}

	return rtClient, nil

}

func AuthenticateRocketChatClient(rtClient RocketChatClient, config Config) error {
	_, errUser := rtClient.Login(&config.Credentials)
	return errUser
}

// SendNotification connects to RocketChat server, authenticate the user and send the notification
func SendNotification(rtClient RocketChatClient, data template.Data) error {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return errRoom
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {
		message := formatMessage(alert)
		_, errMessage := rtClient.SendMessage(channel, message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return errMessage
		}
	}
	return nil
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
