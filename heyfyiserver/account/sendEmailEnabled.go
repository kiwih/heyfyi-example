// +build yesemail

package account

import (
	"bytes"
	"log"
	"net/smtp"
)

func sendEmail(to string, subject string, message string) error {
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		log.Println("Error in sendEmail to:", to, ":", err)
		return err
	}
	defer c.Close()
	// Set the sender and recipient.
	c.Mail("noreply@hey.fyi")
	c.Rcpt(to)
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Println("Error in sendEmail to:", to, ":", err)
		return err
	}
	defer wc.Close()
	buf := bytes.NewBufferString("From: noreply@hey.fyi (hey fyi)\r\nTo: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + message)
	if _, err = buf.WriteTo(wc); err != nil {
		log.Println("Error in sendEmail to:", to, ":", err)
		return err
	}

	log.Println("Successfully sent email to:", to)
	return nil

}
