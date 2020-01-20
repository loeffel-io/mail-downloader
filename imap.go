package main

import (
	i "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	m "github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/charmap"
	"log"
	"strings"
	"unicode/utf8"
)

type imap struct {
	Username string
	Password string
	Server   string
	Port     string
	Client   *client.Client
}

func (imap *imap) connect() error {
	c, err := client.DialTLS(imap.Server+":"+imap.Port, nil)

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
	charset.RegisterEncoding("ansi", charmap.Windows1252)
	charset.RegisterEncoding("iso8859-15", charmap.ISO8859_15)
	i.CharsetReader = charset.Reader
}

func (imap *imap) fixUtf(str string) string {
	callable := func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}

		return r
	}

	return strings.Map(callable, str)
}

func (imap *imap) fetchMessages(mailbox *i.MailboxStatus, mailsChan chan *mail) error {
	seqset := new(i.SeqSet)
	seqset.AddRange(uint32(1), mailbox.Messages)
	messages := make(chan *i.Message)
	section := new(i.BodySectionName)

	go func() {
		if err := imap.Client.Fetch(seqset, []i.FetchItem{section.FetchItem(), i.FetchEnvelope}, messages); err != nil {
			log.Println(err)
		}
	}()

	for message := range messages {
		mail := new(mail)
		mail.fetchMeta(message)

		reader := message.GetBody(section)

		if reader == nil {
			return errors.New("no reader")
		}

		mailReader, err := m.CreateReader(reader)

		if err != nil {
			mail.Error = err
			mailsChan <- mail

			if mailReader != nil {
				if err := mailReader.Close(); err != nil {
					log.Fatal(err)
				}
			}

			continue
		}

		mail.Error = mail.fetchBody(mailReader)
		mailsChan <- mail

		if mailReader != nil {
			if err := mailReader.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}

	return nil
}
