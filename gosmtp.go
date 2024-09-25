package gosmtp

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/sethvargo/go-envconfig"
	gsm "github.com/xhit/go-simple-mail/v2"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Config struct {
	URL string `env:"SMTP_URL, default=smtp://user:pass@localhost:1025"`
}

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
}

var config *Config
var client *gsm.SMTPClient

func Open() {
	config = new(Config)

	// Load config from environment variables
	err := envconfig.Process(context.Background(), config)
	if err != nil {
		panic(err)
	}

	parsed, err := url.Parse(config.URL)
	if err != nil {
		panic(err)
	}

	smtpClient := gsm.NewSMTPClient()

	if parsed.Hostname() != "" {
		smtpClient.Host = parsed.Hostname()
	}

	if parsed.Port() != "" {
		smtpClient.Port, _ = strconv.Atoi(parsed.Port())
	}

	if parsed.User.Username() != "" {
		smtpClient.Username = parsed.User.Username()
	}

	if _, ok := parsed.User.Password(); ok {
		smtpClient.Password, _ = parsed.User.Password()
	}

	if parsed.Scheme == "smtps" {
		smtpClient.Encryption = mail.EncryptionSTARTTLS
	}

	smtpClient.KeepAlive = true

	client, err = smtpClient.Connect()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			client.Noop()
			time.Sleep(30 * time.Second)
		}
	}()
}

func Close() error {
	return client.Close()
}

func Send(mail *Mail) error {
	msg := gsm.NewMSG()
	msg.SetFrom(mail.From)
	msg.AddTo(mail.To)
	msg.SetSubject(mail.Subject)
	msg.SetBody(gsm.TextHTML, mail.Body)

	if err := msg.Send(client); err != nil {
		return err
	}

	return nil
}
