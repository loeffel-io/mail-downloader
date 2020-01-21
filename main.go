package main

import (
	"flag"
	pdf "github.com/adrg/go-wkhtmltopdf"
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
	from := flag.String("from", "", "from")
	to := flag.String("to", "", "to")
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

	// Mailbox
	_, err := imap.selectMailbox("INBOX")

	// search uids
	fromDate, err := time.Parse("2006-01-02", *from) // yyyy-MM-dd ISO 8601

	if err != nil {
		log.Fatal(err)
	}

	toDate, err := time.Parse("2006-01-02", *to) // yyyy-MM-dd ISO 8601

	if err != nil {
		log.Fatal(err)
	}

	uids, err := imap.search(fromDate, toDate)

	if err != nil {
		log.Fatal(err)
	}

	// seqset
	seqset := imap.createSeqSet(uids)

	// channel
	var mailsChan = make(chan *mail)

	// start bar
	bar := pb.StartNew(len(uids))

	// fetch messages
	go func() {
		if err = imap.fetchMessages(seqset, mailsChan); err != nil {
			log.Fatal(err)
		}
	}()

	// pdf
	if err := pdf.Init(); err != nil {
		log.Fatal(err)
	}

	defer pdf.Destroy()

	// out messages
	for mail := range mailsChan {
		if mail.Error != nil {
			log.Println(mail.getErrorText())
			bar.Increment()
			continue
		}

		if len(mail.Attachments) == 0 {
			if err := mail.generateBodyPdf(); err != nil {
				log.Fatal(err)
			}

			bar.Increment()
			continue
		}

		for _, attachment := range mail.Attachments {
			dir := mail.getDirectoryName(imap.Username)

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Println(err)
				continue
			}

			if err = ioutil.WriteFile(dir+"/"+attachment.Filename, attachment.Body, 0644); err != nil {
				log.Println(err)
				continue
			}
		}

		bar.Increment()
	}
}
