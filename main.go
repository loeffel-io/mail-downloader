package main

import (
	"flag"
	"github.com/cheggaaa/pb"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	// flags
	username := flag.String("username", "", "username")
	password := flag.String("password", "", "password")
	server := flag.String("server", "", "server")
	port := flag.String("port", "", "port")
	flag.Parse()

	// imap
	imap := &imap{
		Username: *username,
		Password: *password,
		Server:   *server,
		Port:     *port,
	}

	if err := imap.connect(); err != nil {
		log.Fatal(err)
	}

	if err := imap.login(); err != nil {
		log.Fatal(err)
	}

	imap.enableCharsetReader()

	// mailbox
	inbox, err := imap.getMailbox("INBOX")

	if err != nil {
		log.Fatal(err)
	}

	// search uids
	from := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Now()

	uids, err := imap.search(from, to)

	if err != nil {
		log.Fatal(err)
	}

	// seqset
	seqset := imap.createSeqSet(uids)

	// channel
	var mailsChan = make(chan *mail, inbox.Messages)

	// start bar
	bar := pb.StartNew(len(uids))

	// fetch messages
	go func() {
		if err = imap.fetchMessages(inbox, seqset, mailsChan); err != nil {
			log.Fatal(err)
		}
	}()

	// out messages
	for mail := range mailsChan {
		if mail.Error != nil {
			log.Println(mail.getErrorText())
			bar.Increment()
			continue
		}

		if len(mail.Attachments) == 0 {
			bar.Increment()
			continue
		}

		for _, attachment := range mail.Attachments {
			dir := mail.getDirectoryName(imap.Username)

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Println(err)
				bar.Increment()
				continue
			}

			if err = ioutil.WriteFile(dir+"/"+attachment.Filename, attachment.Body, 0644); err != nil {
				log.Println(err)
				bar.Increment()
				continue
			}
		}
	}
}
