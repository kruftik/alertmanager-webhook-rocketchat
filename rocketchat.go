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
	SendMessage(message *models.Message) (*models.Message, error)
	NewMessage(channel *models.Channel, text string) *models.Message
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

func formatMessage(rtClient RocketChatClient, channel *models.Channel, alert template.Alert) *models.Message {

	const (
		warning       = "warning"
		critical      = "critical"
		warningColor  = "#f2e826"
		criticalColor = "#ef0b1e"
		defaultColor  = "#ffffff"
	)

	var color string
	severity := alert.Labels["severity"]
	title := fmt.Sprintf("**%s: %s**", severity, alert.Annotations["summary"])
	message := rtClient.NewMessage(channel, title)

	if strings.ToLower(severity) == warning {
		color = warningColor
	} else if strings.ToLower(severity) == critical {
		color = criticalColor
	} else {
		color = defaultColor
	}

	var keys []string
	for k := range alert.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	attachementText := fmt.Sprintf("**description**: %s\n", alert.Annotations["description"])
	for _, k := range keys {
		attachementText += fmt.Sprintf("**%s**: %s\n", k, alert.Labels[k])
	}

	message.PostMessage.Attachments = []models.Attachment{
		models.Attachment{
			Color: color,
			Text:  attachementText,
		},
	}

	return message
}

// SendNotification connects to RocketChat server, authenticates the user and sends the notification
func SendNotification(rtClient RocketChatClient, data template.Data) error {

	channelID, errRoom := rtClient.GetChannelId(data.CommonLabels["channel_name"])
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return errRoom
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {

		message := formatMessage(rtClient, channel, alert)
		_, errMessage := rtClient.SendMessage(message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return errMessage
		}
	}
	return nil
}
