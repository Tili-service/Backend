package email

import (
	"log"
	"os"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/resend/resend-go/v2"
	"github.com/wneessen/go-mail"
)

type Sender interface {
	SendEmail(to string, subject string, body string) error
}

type MailpitSender struct {
	client *mail.Client
}

type ResendSender struct {
	client *resend.Client
}

func NewEmailSender() (Sender, error) {
	env := os.Getenv("APP_ENV")

	if env == "production" {
		log.Println("Using Resend for email sending in production environment")
		apiKey := os.Getenv("RESEND_API_KEY")
		client := resend.NewClient(apiKey)
		return &ResendSender{client: client}, nil
	} else {
		log.Println("Using Mailpit for email sending in development environment")
		host := os.Getenv("TILI_SMTP_HOST")
		portStr := os.Getenv("TILI_SMTP_PORT")
		user := os.Getenv("TILI_SMTP_USER")
		pass := os.Getenv("TILI_SMTP_PASS")

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}

		options := []mail.Option{
			mail.WithPort(port),
		}

		if user != "" && pass != "" {
			options = append(options,
				mail.WithSMTPAuth(mail.SMTPAuthPlain),
				mail.WithUsername(user),
				mail.WithPassword(pass),
			)
		}

		if port == 1025 {
			options = append(options, mail.WithTLSPolicy(mail.NoTLS))
		} else {
			options = append(options, mail.WithSSLPort(true))
		}

		client, err := mail.NewClient(host, options...)
		if err != nil {
			return nil, err
		}

		return &MailpitSender{client: client}, nil
	}
}

func GetWelcomeEmailContent(name string, email string) (string, error) {
	tpl, err := pongo2.FromFile("html/welcome_email.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"name":  name,
		"email": email,
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

func GetNewPaymentLinkEmailContent(offer string, paymentLinkURL string) (string, error) {
	tpl, err := pongo2.FromFile("html/new_payment_link.html")
	if err != nil {
		log.Fatalf("Erreur lors de la lecture du template: %s", err)
	}
	ctx := pongo2.Context{
		"offer":            offer,
		"payment_link_url": paymentLinkURL,
	}

	content, err := tpl.Execute(ctx)
	if err != nil {
		log.Fatalf("Erreur lors de l'exécution du template: %s", err)
	}
	return content, nil
}

func (m *MailpitSender) SendEmail(to string, subject string, body string) error {
	msg := mail.NewMsg()
	msg.From("tili-service@local.dev")
	msg.To(to)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, body)
	return m.client.DialAndSend(msg)
}

func (r *ResendSender) SendEmail(to string, subject string, body string) error {
	params := &resend.SendEmailRequest{
		From:    "Tili-Services<no-reply@nduboi.fr>", // Todo change when we have tili domain and verify it with resend(default is domain nduboi.fr)
		To:      []string{to},
		Subject: subject,
		Html:    body,
	}

	r.client.Emails.Send(params)
	return nil
}
