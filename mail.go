package main

import (
	"bytes"
	"fmt"
	pdf "github.com/adrg/go-wkhtmltopdf"
	i "github.com/emersion/go-imap"
	m "github.com/emersion/go-message/mail"
	"github.com/gabriel-vasile/mimetype"
	"github.com/loeffel-io/tax/counter"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type mail struct {
	Uid         uint32
	MessageID   string
	Subject     string
	From        []*i.Address
	Date        time.Time
	Body        [][]byte
	Attachments []*attachment
	Error       error
}

type attachment struct {
	Filename string
	Body     []byte
}

func (mail *mail) fetchMeta(message *i.Message) {
	mail.Uid = message.Uid
	mail.MessageID = message.Envelope.MessageId
	mail.Subject = message.Envelope.Subject
	mail.From = message.Envelope.From
	mail.Date = message.Envelope.Date
}

func (mail *mail) fetchBody(reader *m.Reader) error {
	var (
		bodies      [][]byte
		attachments []*attachment
		count       = counter.CreateCounter()
	)

	for {
		part, err := reader.NextPart()

		if err != nil {
			if err == io.EOF || err.Error() == "multipart: NextPart: EOF" {
				break
			}

			return err
		}

		switch header := part.Header.(type) {
		case *m.InlineHeader:
			body, err := ioutil.ReadAll(part.Body)

			if err != nil {
				if err == io.ErrUnexpectedEOF {
					continue
				}

				return err
			}

			bodies = append(bodies, body)
		case *m.AttachmentHeader:
			// This is an attachment
			filename, err := header.Filename()

			if err != nil {
				return err
			}

			body, err := ioutil.ReadAll(part.Body)

			if err != nil {
				return err
			}

			if filename == "" {
				mime := mimetype.Detect(body)
				filename = fmt.Sprintf("%d-%d%s", count.Next(), mail.Date.Unix(), mime.Extension())
			}

			filename = new(imap).fixUtf(filename)

			attachments = append(attachments, &attachment{
				Filename: filename,
				Body:     body,
			})
		}
	}

	mail.Body = bodies
	mail.Attachments = attachments

	return nil
}

func (mail *mail) generateBodyPdf() error {
	converter, err := pdf.NewConverter()

	if err != nil {
		return err
	}

	defer converter.Destroy()

	converter.Title = "Sample document"
	converter.PaperSize = pdf.A4
	converter.Orientation = pdf.Portrait
	converter.MarginTop = "1cm"
	converter.MarginBottom = "1cm"
	converter.MarginLeft = "10mm"
	converter.MarginRight = "10mm"

	for _, body := range mail.Body {
		object, err := pdf.NewObjectFromReader(bytes.NewReader(body))

		if err != nil {
			return err
		}

		converter.Add(object)
	}

	outFile, err := os.Create(fmt.Sprintf("mail-%d.pdf", time.Now().Unix()))

	if err != nil {
		return err
	}

	if err := converter.Run(outFile); err != nil {
		return err
	}

	return outFile.Close()
}

func (mail *mail) getDirectoryName(username string) string {
	return fmt.Sprintf(
		"files/%s/%s-%d/%s",
		username, mail.Date.Month(), mail.Date.Year(), mail.From[0].HostName,
	)
}

func (mail *mail) getErrorText() string {
	return fmt.Sprintf("Error: %s\nSubject: %s\nFrom: %s\n", mail.Error.Error(), mail.Subject, mail.Date.Local())
}
