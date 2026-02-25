package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

func SendMail(host string, port int, user, password, fromName string, ssl bool, to, subject, body string) error {
	from := user
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, user)
	}

	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", user, password, host)

	if ssl {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS连接失败: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("SMTP客户端创建失败: %w", err)
		}
		defer client.Close()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}

		if err := client.Mail(user); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}

		w, err := client.Data()
		if err != nil {
			return err
		}
		w.Write([]byte(msg))
		w.Close()

		return client.Quit()
	}

	return smtp.SendMail(addr, auth, user, []string{to}, []byte(msg))
}
