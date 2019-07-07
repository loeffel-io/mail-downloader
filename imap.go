package main

import (
	"github.com/cheggaaa/pb"
	i "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	m "github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
	"log"
)

type imap struct {
	Username string
	Password string
	Client   *client.Client
}

func (imap *imap) connect() error {
	c, err := client.DialTLS("imap.gmail.com:993", nil)

	if err != nil {
		return err
	}

	imap.Client = c
	return nil
}

func (imap *imap) login() error {
	return imap.Client.Login(imap.Username, imap.Password)
}

func (imap *imap) getMailbox(mailbox string) (*i.MailboxStatus, error) {
	return imap.Client.Select(mailbox, true)
}

func (imap *imap) enableCharsetReader() {
	i.CharsetReader = charset.Reader
}

func (imap *imap) fetchMessages(mailbox *i.MailboxStatus, bar *pb.ProgressBar) ([]*mail, error) {
	var mails []*mail
	seqset := new(i.SeqSet)
	seqset.AddRange(1, mailbox.Messages)
	messages := make(chan *i.Message)
	section := new(i.BodySectionName)

	go func() {
		if err := imap.Client.Fetch(seqset, []i.FetchItem{section.FetchItem(), i.FetchEnvelope}, messages); err != nil {
			log.Fatal(err)
		}
	}()

	for message := range messages {
		mail := new(mail)
		mail.fetchMeta(message)

		reader := message.GetBody(section)

		if reader == nil {
			return nil, errors.New("no reader")
		}

		mailReader, err := m.CreateReader(reader)

		if err != nil {
			mail.Error = err
			mails = append(mails, mail)
			mailReader.Close()
			bar.Increment()
			continue
		}

		mail.Error = mail.fetchBody(mailReader)
		mails = append(mails, mail)
		mailReader.Close()
		bar.Increment()
	}

	return mails, nil
}
