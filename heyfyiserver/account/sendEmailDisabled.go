// +build !yesemail

package account

import "log"

func sendEmail(to string, subject string, message string) error {
	log.Printf("Imaginary Email:\r\nTo: %s\r\nSubject: %s\r\nMessage: %s\r\n\r\n", to, subject, message)
	return nil

}
