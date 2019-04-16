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

type ChannelInfo struct {
	DefaultChannelName string
}

type Config struct {
	Endpoint       url.URL
	Credentials    models.UserCredentials
	SeverityColors map[string]string
	Channel        ChannelInfo
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

func formatMessage(rtClient RocketChatClient, channel *models.Channel, alert template.Alert, config Config) *models.Message {

	const (
		defaultColor = "#ffffff"
	)

	severity := alert.Labels["severity"]
	title := fmt.Sprintf("**%s: %s**", severity, alert.Annotations["summary"])
	message := rtClient.NewMessage(channel, title)

	color := defaultColor
	for k, v := range config.SeverityColors {
		if k == strings.ToLower(severity) {
			color = v
		}
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
func SendNotification(rtClient RocketChatClient, data template.Data, config Config) error {

	var channelName string
	if val, ok := data.CommonLabels["channel_name"]; ok {
		channelName = val
	} else {
		channelName = config.Channel.DefaultChannelName
	}

	channelID, errRoom := rtClient.GetChannelId(channelName)
	if errRoom != nil {
		log.Printf("Error to get room ID: %v", errRoom)
		return errRoom
	}
	channel := &models.Channel{ID: channelID}

	log.Printf("Alerts: Status=%s, GroupLabels=%v, CommonLabels=%v", data.Status, data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {

		message := formatMessage(rtClient, channel, alert, config)
		_, errMessage := rtClient.SendMessage(message)
		if errMessage != nil {
			log.Printf("Error to send message: %v", errMessage)
			return errMessage
		}
	}
	return nil
}
