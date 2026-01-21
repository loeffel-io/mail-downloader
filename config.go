package main

type Config struct {
	Imap struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
	} `yaml:"imap"`

	Attachments map[string]struct {
		Subjects []string `yaml:"subjects"`
	} `yaml:"attachments"`

	Mails struct {
		Subjects []string `yaml:"subjects"`
	} `yaml:"mails"`
}
