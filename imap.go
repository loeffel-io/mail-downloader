package main

import (
	i "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	m "github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
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

func (imap *imap) fetchMessages(mailbox *i.MailboxStatus) ([]*mail, error) {
	seqset := new(i.SeqSet)
	seqset.AddRange(mailbox.Messages, mailbox.Messages-4000)
	messages := make(chan *i.Message, 4000+1)
	section := new(i.BodySectionName)

	if err := imap.Client.Fetch(seqset, []i.FetchItem{section.FetchItem(), i.FetchEnvelope}, messages); err != nil {
		return nil, err
	}

	var mails []*mail
	for message := range messages {
		reader := message.GetBody(section)

		if reader == nil {
			return nil, errors.New("no reader")
		}

		mailReader, err := m.CreateReader(reader)

		if err != nil {
			return nil, err
		}

		mails = append(mails, imap.parseMail(message, mailReader))
	}

	return mails, nil
}

func (imap *imap) parseMail(message *i.Message, mailReader *m.Reader) *mail {
	defer mailReader.Close()

	mail := new(mail)
	mail.fetchMeta(message)
	mail.Error = mail.fetchBody(mailReader)

	return mail
}
