package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// EmailConfig holds the configuration for sending emails
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SenderEmail  string
	SenderName   string
	SMTPPassword string
}

// NotificationService handles sending notifications to users
type NotificationService struct {
	Config EmailConfig
}

// NewNotificationService creates a new notification service
func NewNotificationService(config EmailConfig) *NotificationService {
	return &NotificationService{
		Config: config,
	}
}

// SendWeatherAlert sends a weather alert email to the user
func (s *NotificationService) SendWeatherAlert(to, city, condition, severity string) error {
	// Format the email
	subject := fmt.Sprintf("Weather Alert: %s conditions in %s", severity, city)
	body := fmt.Sprintf(`
Hello,

We're sending you this alert about current weather conditions in %s.

Current Condition: %s
Severity: %s

Please take necessary precautions and stay safe.

Best regards,
Go Weather Team
`, city, condition, severity)

	// Send the email
	return s.sendEmail(to, subject, body)
}

// SendDailyReport sends a daily temperature report for a user's home city
func (s *NotificationService) SendDailyReport(to, city string, temp float64, condition string, forecasts []interface{}) error {
	// Format the email
	subject := fmt.Sprintf("Daily Weather Report for %s", city)

	// Build email body
	body := fmt.Sprintf(`
Hello,

Here is your daily weather report for %s:

Current Temperature: %.1f°C
Current Conditions: %s

`, city, temp, condition)

	// Add forecast information if available
	if len(forecasts) > 0 {
		body += "\nForecast for the upcoming days:\n\n"

		for i, forecast := range forecasts {
			if i >= 5 { // Limit to 5 days
				break
			}

			f, ok := forecast.(map[string]interface{})
			if !ok {
				continue
			}

			day, _ := f["day"].(string)
			date, _ := f["date"].(string)
			maxTemp, _ := f["maxTemp"].(float64)
			minTemp, _ := f["minTemp"].(float64)
			description, _ := f["description"].(string)

			body += fmt.Sprintf("%s (%s): %.1f°C to %.1f°C, %s\n",
				day, date, minTemp, maxTemp, description)
		}
	}

	body += `

We'll keep you updated with daily reports and alerts for any significant changes.

Best regards,
Go Weather Team
`

	// Send the email
	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using the configured SMTP server
func (s *NotificationService) sendEmail(to, subject, body string) error {
	// Format the email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.Config.SenderName, s.Config.SenderEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=utf-8"

	// Compose the message
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Setup authentication
	auth := smtp.PlainAuth("", s.Config.SenderEmail, s.Config.SMTPPassword, s.Config.SMTPHost)

	// Send the email
	err := smtp.SendMail(
		s.Config.SMTPHost+":"+s.Config.SMTPPort,
		auth,
		s.Config.SenderEmail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	log.Printf("Successfully sent email to %s", to)
	return nil
}

// IsSevereWeather determines if the weather condition is severe enough based on user's alert threshold
func IsSevereWeather(condition string, threshold string) bool {
	condition = strings.ToLower(condition)

	// Define severe conditions
	severeConditions := []string{"thunderstorm", "storm", "hurricane", "tornado", "blizzard", "heavy rain", "heavy snow", "flood"}
	moderateConditions := append(severeConditions, "rain", "snow", "fog", "mist", "drizzle", "wind")

	switch threshold {
	case "severe":
		for _, c := range severeConditions {
			if strings.Contains(condition, c) {
				return true
			}
		}
	case "moderate":
		for _, c := range moderateConditions {
			if strings.Contains(condition, c) {
				return true
			}
		}
	case "all":
		return true
	}

	return false
}

// Helper function to get environment variables with fallback to default values
func getEnv(key, fallback string) string {
	value := fallback
	if envVal, exists := os.LookupEnv(key); exists {
		value = envVal
	}
	return value
}
