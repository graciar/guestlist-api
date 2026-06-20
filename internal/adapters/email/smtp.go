package email

import (
	"fmt"

	"github.com/graciar/guestlist-api/internal/env"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

func SendEmail(recipientEmail, recipientName, subject, textPart, htmlPart string) error {
	// 1. Fetch env variables inside the function to ensure they are read correctly at runtime
	publicKey := env.GetString("MJ_APIKEY_PUBLIC", "")
	privateKey := env.GetString("MJ_APIKEY_PRIVATE", "")
	senderEmail := env.GetString("MJ_SENDER_EMAIL", "")
	senderName := env.GetString("MJ_SENDER_NAME", "")

	if publicKey == "" || privateKey == "" {
		return fmt.Errorf("mailjet API keys are not set in environment variables")
	}

	// 2. Initialize the Mailjet client
	mj := mailjet.NewMailjetClient(publicKey, privateKey)

	// 3. Construct the email message mapping
	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: senderEmail,
					Name:  senderName,
				},
				To: &mailjet.RecipientsV31{
					{
						Email: recipientEmail,
						Name:  recipientName,
					},
				},
				Subject:  subject,
				TextPart: textPart,
				HTMLPart: htmlPart,
			},
		},
	}

	// 4. Send the message
	res, err := mj.SendMailV31(&messages)
	if err != nil {
		return fmt.Errorf("failed to send email via Mailjet: %w", err)
	}

	// 5. Check response status safely
	if res != nil && len(res.ResultsV31) > 0 {
		if res.ResultsV31[0].Status != "success" {
			return fmt.Errorf("mailjet returned status: %s", res.ResultsV31[0].Status)
		}
	} else {
		return fmt.Errorf("mailjet returned an empty response payload")
	}

	return nil
}
