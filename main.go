package main

import (
	"io/ioutil"
	"log"
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

	// mailbox
	inbox, err := imap.getMailbox("INBOX")

	if err != nil {
		log.Fatal(err)
	}

	mails, err := imap.fetchMessages(inbox)

	if err != nil {
		log.Fatal(err)
	}

	for _, mail := range mails {
		if len(mail.Attachments) == 0 {
			continue
		}

		for _, attachment := range mail.Attachments {
			err = ioutil.WriteFile("files/"+attachment.Filename, attachment.Body, 0644)

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
