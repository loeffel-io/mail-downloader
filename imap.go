package main

import (
	i "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	m "github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"time"
)

type imap struct {
	Username string
	Password string
	Client   *client.Client
}

type mail struct {
	Subject     string
	From        []*m.Address
	Date        time.Time
	Text        [][]byte
	Attachments []*attachment
}

type attachment struct {
	Filename string
	Body     []byte
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

func (imap *imap) fetchMessages(mailbox *i.MailboxStatus) ([]*mail, error) {
	seqset := new(i.SeqSet)
	seqset.AddRange(mailbox.Messages, mailbox.Messages-100)
	messages := make(chan *i.Message, 100+1)
	section := new(i.BodySectionName)

	if err := imap.Client.Fetch(seqset, []i.FetchItem{section.FetchItem()}, messages); err != nil {
		return nil, err
	}

	var mails []*mail
	for message := range messages {
		reader := message.GetBody(section)

		if reader == nil {
			return nil, errors.New("no message body")
		}

		mailReader, err := m.CreateReader(reader)

		if err != nil {
			return nil, err
		}

		mail, err := imap.readMessage(mailReader)

		if err != nil {
			return nil, err
		}

		mails = append(mails, mail)
	}

	return mails, nil
}

func (imap *imap) readMessage(reader *m.Reader) (*mail, error) {
	subject, err := reader.Header.Subject()

	if err != nil {
		return nil, err
	}

	from, err := reader.Header.AddressList("From")

	if err != nil {
		return nil, err
	}

	date, err := reader.Header.Date()

	if err != nil {
		return nil, err
	}

	var mailTexts [][]byte
	var mailAttachments []*attachment

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			continue
		}

		switch header := part.Header.(type) {
		case *m.InlineHeader:
			body, err := ioutil.ReadAll(part.Body)

			if err != nil {
				return nil, err
			}

			mailTexts = append(mailTexts, body)
		case *m.AttachmentHeader:
			// This is an attachment
			filename, err := header.Filename()

			if err != nil {
				return nil, err
			}

			body, err := ioutil.ReadAll(part.Body)

			if err != nil {
				return nil, err
			}

			mailAttachments = append(mailAttachments, &attachment{
				Filename: filename,
				Body:     body,
			})
		}
	}

	return &mail{
		Subject:     subject,
		From:        from,
		Date:        date,
		Text:        mailTexts,
		Attachments: mailAttachments,
	}, nil
}
