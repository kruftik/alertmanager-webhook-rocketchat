package main

import (
	"testing"

	"github.com/RocketChat/Rocket.Chat.Go.SDK/models"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

type FormatMessageOK struct {
	input    template.Alert
	err      error
	expected *models.Message
}

var formatMessagesOK = []FormatMessageOK{
	{
		input: template.Alert{},
		expected: &models.Message{
			ID:     "0.64464",
			RoomID: "1",
			Msg:    "**: **",
			PostMessage: models.PostMessage{
				Attachments: []models.Attachment{
					models.Attachment{
						Color: "#ffffff",
						Text:  "",
					},
				},
			},
		},
	},
	{
		input: template.Alert{
			Status: "warning",
			Labels: template.KV{
				"alertname":    "CPUHigh",
				"channel_name": "prometheus-test-room",
				"exported_job": "aws_ec2",
				"instance":     "localhost:9106",
				"instance_id":  "i-06d34d52562f5ac45",
				"job":          "cloudwatch",
				"severity":     "high",
			},
			Annotations: template.KV{
				"description": "EC2 CPU Utilization on instance i-06d34d52562f5ac45 is 2.4918032786885242%",
				"summary":     "Instance ID i-06d34d52562f5ac45 above threshold",
			},
		},
		expected: &models.Message{
			ID:     "0.64464",
			RoomID: "1",
			Msg:    "**Warning: Instance ID i-06d34d52562f5ac45 above threshold**",
			PostMessage: models.PostMessage{
				Attachments: []models.Attachment{
					models.Attachment{
						Color: "#f2e826",
						Text:  "",
					},
				},
			},
		},
	},
	{
		input: template.Alert{
			Status: "critical",
			Labels: template.KV{
				"alertname":    "CPUHigh",
				"channel_name": "prometheus-test-room",
				"exported_job": "aws_ec2",
				"instance":     "localhost:9106",
				"instance_id":  "i-06d34d52562f5ac45",
				"job":          "cloudwatch",
				"severity":     "high",
			},
			Annotations: template.KV{
				"description": "EC2 CPU Utilization on instance i-06d34d52562f5ac45 is 2.4918032786885242%",
				"summary":     "Instance ID i-06d34d52562f5ac45 above threshold",
			},
		},
		expected: &models.Message{
			ID:     "0.64464",
			RoomID: "1",
			Msg:    "**Critical: Instance ID i-06d34d52562f5ac45 above threshold**",
			PostMessage: models.PostMessage{
				Attachments: []models.Attachment{
					models.Attachment{
						Color: "#ef0b1e",
						Text:  "",
					},
				},
			},
		},
	},
	{
		input: template.Alert{
			Status: "critic",
			Labels: template.KV{
				"alertname":    "CPUHigh",
				"channel_name": "prometheus-test-room",
				"exported_job": "aws_ec2",
				"instance":     "localhost:9106",
				"instance_id":  "i-06d34d52562f5ac45",
				"job":          "cloudwatch",
				"severity":     "high",
			},
			Annotations: template.KV{
				"description": "EC2 CPU Utilization on instance i-06d34d52562f5ac45 is 2.4918032786885242%",
				"summary":     "Instance ID i-06d34d52562f5ac45 above threshold",
			},
		},
		expected: &models.Message{
			ID:     "0.64464",
			RoomID: "1",
			Msg:    "**Critic: Instance ID i-06d34d52562f5ac45 above threshold**",
			PostMessage: models.PostMessage{
				Attachments: []models.Attachment{
					models.Attachment{
						Color: "#ffffff",
						Text:  "",
					},
				},
			},
		},
	},
}

func TestFormatMessage(t *testing.T) {

	channel := &models.Channel{ID: "1"}
	var message *models.Message

	for _, data := range formatMessagesOK {
		message = formatMessage(channel, data.input)
		assert.Equal(t, message.Msg, data.expected.Msg)
		assert.Equal(t, message.Attachments[0].Color, data.expected.Attachments[0].Color)
	}
}
