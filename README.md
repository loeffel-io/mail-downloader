# Mail Attachment Downloader

- wkhtmltopdf required (!)

```bash
# date format: yyyy-MM-dd - ISO 8601 

make install
go build -o mail-attachment-downloader
./mail-attachment-downloader -config=config.yml -from="2019-10-01" -to="2019-12-31"
```

### Config

```yaml
imap:
  username: secret@gmail.com
  password: secret
  server: imap.gmail.com
  port: 993

attachments:
  mimetypes:
    - application/pdf

mails:
  subjects:
    - invoice
    - rechnung
    - receipt
```