package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	imap := &imap{
		Username: "lucas.loeffel@gmail.com",
		Password: "eoppezuwhutdhxpb",
	}

	// connection
	if err := imap.connect(); err != nil {
		log.Fatal(err)
	}

	// login
	if err := imap.login(); err != nil {
		log.Fatal(err)
	}

	// charset reader
	imap.enableCharsetReader()

	// mailbox
	inbox, err := imap.getMailbox("INBOX")

	if err != nil {
		log.Fatal(err)
	}

	// fetch messages
	mails, err := imap.fetchMessages(inbox)

	if err != nil {
		log.Fatal(err)
	}

	// out messages
	for _, mail := range mails {
		if mail.Error != nil {
			log.Printf(
				"ERROR: %s\nSUBJECT: %s\nFROM: %s\n",
				mail.Error.Error(), mail.Subject, mail.Date.Local(),
			)
			continue
		}

		if len(mail.Attachments) == 0 {
			continue
		}

		for _, attachment := range mail.Attachments {
			dir := fmt.Sprintf(
				"files/%s/%s-%d/%s",
				imap.Username, mail.Date.Month(), mail.Date.Year(), mail.From[0].HostName,
			)

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile(dir+"/"+attachment.Filename, attachment.Body, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}
