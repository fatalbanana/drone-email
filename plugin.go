package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aymerick/douceur/inliner"
	"github.com/drone/drone-go/template"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	gomail "github.com/go-mail/mail"
	"github.com/jaytaylor/html2text"
)

type (
	Repo struct {
		FullName string
		Owner    string
		Name     string
		SCM      string
		Link     string
		Avatar   string
		Branch   string
		Private  bool
		Trusted  bool
	}

	Remote struct {
		URL string
	}

	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	Commit struct {
		Sha     string
		Ref     string
		Branch  string
		Link    string
		Message string
		Author  Author
	}

	Build struct {
		Number   int
		Event    string
		Status   string
		Link     string
		Created  int64
		Started  int64
		Finished int64
	}

	PrevBuild struct {
		Status string
		Number int
	}

	PrevCommit struct {
		Sha string
	}

	Prev struct {
		Build  PrevBuild
		Commit PrevCommit
	}

	Job struct {
		Status   string
		ExitCode int
		Started  int64
		Finished int64
	}

	Yaml struct {
		Signed   bool
		Verified bool
	}

	Config struct {
		FromAddress    string
		FromName       string
		Host           string
		Port           int
		Username       string
		Password       string
		SkipVerify     bool
		NoStartTLS     bool
		Recipients     []string
		RecipientsFile string
		RecipientsOnly bool
		Subject        string
		Body           string
		Attachment     string
		Attachments    []string
		ClientHostname string
	}

	Plugin struct {
		BuildContext
		Config Config
	}

	BuildContext struct {
		Repo        Repo
		Remote      Remote
		Commit      Commit
		Build       Build
		Prev        Prev
		Job         Job
		Yaml        Yaml
		Tag         string
		PullRequest int
		DeployTo    string
	}
)

var (
	ErrDroneSanity = errors.New("failed sanity check: no finish time for build")
)

func (p Plugin) prepareMessage() (*gomail.Message, error) {
	// Render body in HTML and plain text
	renderedBody, err := template.RenderTrim(p.Config.Body, p.BuildContext)
	if err != nil {
		return nil, fmt.Errorf("could not render body template: %v", err)
	}
	html, err := inliner.Inline(renderedBody)
	if err != nil {
		return nil, fmt.Errorf("could not inline rendered body: %v", err)
	}
	plainBody, err := html2text.FromString(html)
	if err != nil {
		return nil, fmt.Errorf("could not convert html to text: %v", err)
	}

	// Render subject
	subject, err := template.RenderTrim(p.Config.Subject, p.BuildContext)
	if err != nil {
		return nil, fmt.Errorf("could not render subject template: %v", err)
	}

	message := gomail.NewMessage()
	message.SetAddressHeader("From", p.Config.FromAddress, p.Config.FromName)
	message.SetHeader("To", strings.Join(p.Config.Recipients, ", "))
	message.SetHeader("Subject", subject)
	message.AddAlternative("text/plain", plainBody)
	message.AddAlternative("text/html", html)

	if p.Config.Attachment != "" {
		message.Attach(p.Config.Attachment)
	}

	for _, attachment := range p.Config.Attachments {
		message.Attach(attachment)
	}

	return message, nil
}

// Exec will send emails over SMTP
func (p Plugin) Exec() error {

	if p.Build.Finished == 0 {
		return ErrDroneSanity
	}

	if !p.Config.RecipientsOnly && p.Commit.Author.Email != "" {
		exists := false
		for _, recipient := range p.Config.Recipients {
			if recipient == p.Commit.Author.Email {
				exists = true
			}
		}

		if !exists {
			p.Config.Recipients = append(p.Config.Recipients, p.Commit.Author.Email)
		}
	}

	if p.Config.RecipientsFile != "" {
		f, err := os.Open(p.Config.RecipientsFile)
		if err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				p.Config.Recipients = append(p.Config.Recipients, scanner.Text())
			}
		} else {
			return fmt.Errorf("could not open RecipientsFile %s: %v", p.Config.RecipientsFile, err)
		}
	}

	// Send emails
	message, err := p.prepareMessage()
	if err != nil {
		return err
	}
	defer message.Reset()

	client, err := smtp.Dial(fmt.Sprintf("%s:%d", p.Config.Host, p.Config.Port))
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}

	if !p.Config.NoStartTLS {
		tlsConfig := &tls.Config{}
		if p.Config.SkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		err := client.StartTLS(tlsConfig)
		if err != nil {
			return fmt.Errorf("starttls failed: %v", err)
		}
	}

	if p.Config.Username != "" && p.Config.Password != "" {
		auth := sasl.NewPlainClient("", p.Config.Username, p.Config.Password)
		authErr := client.Auth(auth)
		if authErr != nil {
			return fmt.Errorf("auth failed: %v", err)
		}
	}

	err = client.Mail(p.Config.FromAddress, nil)
	if err != nil {
		return fmt.Errorf("error at mail from: %v", err)
	}

	for _, rcpt := range p.Config.Recipients {
		err = client.Rcpt(rcpt, nil)
		if err != nil {
			return fmt.Errorf("error at rcpt(%s) phase: %v", rcpt, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("error before DATA phase: %v", err)
	}

	_, err = message.WriteTo(writer)
	if err != nil {
		return fmt.Errorf("error writing body: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error after DATA: %v", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("error on quit: %v", err)
	}
	return nil
}
