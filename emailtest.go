package main

import (
	"fmt"
	"net/smtp"
)

func main() {
	// Configuration
	from := "klucse2000030442@gmail.com"
	password := "odwbtxvdwbxbslwb" // Your App Password
	to := []string{"knsharshavardhan2003@gmail.com"}
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Compose message
	subject := "Test Email from Go"
	body := "This is a test email sent from Go."
	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to[0], subject, body)

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
	if err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		return
	}

	fmt.Println("Email sent successfully!")
}
