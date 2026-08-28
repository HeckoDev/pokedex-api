package mailer

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/smtp"
	"os"
	"strings"
)

// GenerateCode returns a random numeric verification code of the given length.
func GenerateCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			// Extremely unlikely; fall back to a fixed digit rather than panic.
			code[i] = '0'
			continue
		}
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// SendVerificationCode sends the verification code to the given address.
//
// If no SMTP host is configured (SMTP_HOST empty), it falls back to a
// development mode that simply logs the code to stdout, so the flow works
// locally without any real mail server.
func SendVerificationCode(toEmail, code string) error {
	host := os.Getenv("SMTP_HOST")

	subject := "Pokédex — Code de vérification"
	body := fmt.Sprintf(
		"Bonjour,\r\n\r\nVotre code de vérification Pokédex est : %s\r\n\r\n"+
			"Ce code expire dans 15 minutes.\r\n\r\n"+
			"Si vous n'êtes pas à l'origine de cette demande, ignorez cet email.\r\n",
		code,
	)

	if host == "" {
		// Development fallback — no SMTP configured.
		log.Printf("[mailer:dev] No SMTP configured. Verification code for %s: %s", toEmail, code)
		return nil
	}

	port := getEnv("SMTP_PORT", "587")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := getEnv("SMTP_FROM", user)

	msg := buildMessage(from, toEmail, subject, body)
	addr := fmt.Sprintf("%s:%s", host, port)

	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, password, host)
	}

	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
		log.Printf("[mailer] Failed to send email to %s: %v", toEmail, err)
		return err
	}

	log.Printf("[mailer] Verification email sent to %s", toEmail)
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}
