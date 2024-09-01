package backend

import (
	"fmt"
	"github.com/cocreators-ee/praga"
	"github.com/mailjet/mailjet-apiv3-go/v4"
	"log"
)

var verificationTemplateBytes, _ = praga.VerificationEmail.ReadFile("email/verification.html")
var verificationTemplate = string(verificationTemplateBytes[:])

const textTemplate = "You can login to %s with the code %s or in case of issues contact %s for assistance."

type MailjetSender struct {
	client   *mailjet.Client
	from     string
	fromName string
	subject  string
}

func getMailjetSender(srv *Server) *MailjetSender {
	return &MailjetSender{
		client:   mailjet.NewMailjetClient(srv.Config.Mailjet.APIKeyPublic, srv.Config.Mailjet.APIKeyPrivate),
		from:     srv.Config.Email.From,
		fromName: srv.Config.Email.FromName,
		subject:  fmt.Sprintf("%s verification code", srv.Config.Brand),
	}
}

func (ms MailjetSender) sendEmailViaMailjet(email, brand, code, support string) {
	variables := map[string]interface{}{
		"brand":   brand,
		"code":    code,
		"support": support,
	}

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: ms.from,
				Name:  ms.fromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
				},
			},
			Subject:          ms.subject,
			TextPart:         fmt.Sprintf(textTemplate, brand, code, support),
			HTMLPart:         verificationTemplate,
			TemplateLanguage: true,
			Variables:        variables,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := ms.client.SendMailV31(&messages)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Data: %+v\n", res)
}
