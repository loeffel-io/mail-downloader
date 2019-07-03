package main

import (
	"github.com/loeffel-io/helper"
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

		helper.Debug(mail.Subject, mail.Date.UTC().Local(), mail.From[0].Address)

		for _, attachment := range mail.Attachments {
			if err := os.MkdirAll("files/"+mail.From[0].Address, os.ModePerm); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile("files/"+mail.From[0].Address+"/"+attachment.Filename, attachment.Body, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}
