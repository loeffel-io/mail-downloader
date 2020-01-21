package main

import (
	"flag"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/loeffel-io/tax/search"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	var config *Config

	// flags
	configPath := flag.String("config", "", "config path")
	from := flag.String("from", "", "from date")
	to := flag.String("to", "", "to date")
	flag.Parse()

	// yaml
	yamlBytes, err := ioutil.ReadFile(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	// yaml to config
	err = yaml.Unmarshal(yamlBytes, &config)

	if err != nil {
		log.Fatal(err)
	}

	// imap
	imap := &imap{
		Username: config.Imap.Username,
		Password: config.Imap.Password,
		Server:   config.Imap.Server,
		Port:     config.Imap.Port,
	}

	if err := imap.connect(); err != nil {
		log.Fatal(err)
	}

	if err := imap.login(); err != nil {
		log.Fatal(err)
	}

	imap.enableCharsetReader()

	// Mailbox
	_, err = imap.selectMailbox("INBOX")

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

	// out messages
	for mail := range mailsChan {
		dir := mail.getDirectoryName(imap.Username)

		if mail.Error != nil {
			log.Println(mail.getErrorText())
			bar.Increment()
			continue
		}

		// attachments
		for _, attachment := range mail.Attachments {
			s := &search.Search{
				Search: config.Attachments.Mimetypes,
				Data:   attachment.Mimetype,
			}

			if !s.Find() {
				continue
			}

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Fatal(err)
			}

			if err = ioutil.WriteFile(fmt.Sprintf("%s/%s", dir, attachment.Filename), attachment.Body, 0644); err != nil {
				log.Fatal(err)
			}
		}

		// pdf
		s := &search.Search{
			Search: config.Mails.Subjects,
			Data:   mail.Subject,
		}

		if !s.Find() {
			bar.Increment()
			continue
		}

		bytes, err := mail.generatePdf()

		if err != nil {
			log.Println(err.Error())
			bar.Increment()
			continue
		}

		if bytes == nil {
			bar.Increment()
			continue
		}

		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		if err = ioutil.WriteFile(fmt.Sprintf("%s/mail-%d.pdf", dir, mail.Uid), bytes, 0644); err != nil {
			log.Fatal(err)
		}

		bar.Increment()
	}
}
