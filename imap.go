package main

import (
	"bytes"
	i "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	m "github.com/emersion/go-message/mail"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

type imap struct {
	Username string
	Password string
	Server   string
	Port     string
	Client   *imapclient.Client
}

func (imap *imap) connect() error {
	c, err := imapclient.DialTLS(imap.Server+":"+imap.Port, nil)

	if err != nil {
		return err
	}

	imap.Client = c
	return nil
}

func (imap *imap) login() error {
	return imap.Client.Login(imap.Username, imap.Password).Wait()
}

func (imap *imap) selectMailbox(mailbox string) (*i.SelectData, error) {
	return imap.Client.Select(mailbox).Wait()
}

func (imap *imap) search(from, to time.Time) (*i.SearchData, error) {
	return imap.Client.UIDSearch(&i.SearchCriteria{
		Since:  from,
		Before: to,
	}, nil).Wait()
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

func (imap *imap) fetchMessages(seqset i.SeqSet, mailsChan chan *mail) error {
	msgBuffer, err := imap.Client.UIDFetch(seqset, []i.FetchItem{i.FetchItemBody, i.FetchItemBodyStructure, i.FetchItemEnvelope}).Collect()

	if err != nil {
		log.Println(err)
	}

	for _, message := range msgBuffer {
		mail := new(mail)
		if err = mail.fetchMeta(message); err != nil {
			return err
		}

		var reader *bytes.Reader
		log.Printf("%+v", message.BodyStructure)
		for test, b := range message.BodySection {
			log.Printf("%+v", test)
			return nil
			reader = bytes.NewReader(b)
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
