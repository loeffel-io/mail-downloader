package main

import (
	"github.com/cheggaaa/pb"
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

	// start bar
	bar := pb.StartNew(int(inbox.Messages))

	// fetch messages
	mails, err := imap.fetchMessages(inbox, bar)

	if err != nil {
		log.Fatal(err)
	}

	// stop bar
	bar.Finish()

	// out messages
	for _, mail := range mails {
		if mail.Error != nil {
			log.Println(mail.getErrorText())
			continue
		}

		if len(mail.Attachments) == 0 {
			continue
		}

		for _, attachment := range mail.Attachments {
			dir := mail.getDirectoryName(imap.Username)

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile(dir+"/"+attachment.Filename, attachment.Body, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}
