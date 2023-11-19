package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed email-templates
var tempFS embed.FS

// create template data for email
type EmailTD struct {
	Link string
}

// crete function to send email
func (app *application) SendEmail(to, from, subject, tmpl string, attachments []string, data interface{}) error {
	// create html template
	htmlTemp, err := RenderHtmlTemplate(tmpl, data)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when creating html template : ", err)
		return err
	}

	// create plain template
	plainTemp, err := RenderPlainTemplate(tmpl, data)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when creating plain template : ", err)
		return err
	}

	// create variable for sendirng email credential
	var fromAddressEmail string

	// check on email
	if app.config.mail.FromAddress == "" {
		fromAddressEmail = "test@test.com"
	} else {
		fromAddressEmail = app.config.mail.FromAddress
	}

	// print info
	app.infoLog.Println(htmlTemp, plainTemp)

	// create smtp server
	getServer := mail.NewSMTPClient()

	// configure server email
	getServer.Host = app.config.mail.Host
	getServer.Port = app.config.mail.Port
	getServer.Username = app.config.mail.Username
	getServer.Password = app.config.mail.Password
	//getServer.Encryption = CreateEncryption(app.config.mail.Encryption)
	getServer.Encryption = mail.EncryptionTLS
	getServer.KeepAlive = false
	getServer.ConnectTimeout = 10 * time.Second
	getServer.SendTimeout = 10 * time.Second

	// create smtp client
	smtpClient, err := getServer.Connect()

	// check for an error
	if err != nil {
		log.Println("error when creating cmtpo client : ", err)
		app.errorLog.Println(err)
		return err
	}

	// create mail opbjhect
	mailObj := mail.NewMSG()

	// set attribute of email
	mailObj.SetFrom(fromAddressEmail)
	mailObj.AddTo(to)
	mailObj.SetSubject(subject)
	mailObj.SetBody(mail.TextHTML, htmlTemp)
	mailObj.AddAlternative(mail.TextPlain, plainTemp)

	// check for attachment
	if len(attachments) > 0 {
		for _, att := range attachments {
			// add attachment if attachment exist
			mailObj.AddAttachment(att)
		}
	}

	// send email
	err = mailObj.Send(smtpClient)

	app.infoLog.Println("success sending email")

	if err != nil {
		log.Println("error when sending email : ", err)
		app.errorLog.Println(err)
		return err
	}

	return nil
}

// create function to render html template
func RenderHtmlTemplate(temp string, data interface{}) (string, error) {
	// create path to template
	pathToTemplate := fmt.Sprintf("email-templates/%s.html.tmpl", temp)

	// create template file
	tc, err := template.New("email-template").ParseFS(tempFS, pathToTemplate)

	// check for an error
	if err != nil {
		log.Println("error when creating template html : ", err)
		return "", err
	}

	// create byte to hold data from template\
	var holdData bytes.Buffer

	// execute template
	err = tc.ExecuteTemplate(&holdData, "body", data)

	// cehck for an error
	if err != nil {
		log.Println("error when creating template html with data : ", err)
		return "", err
	}

	// convert byte into string
	result := holdData.String()

	// if okay
	return result, nil
}

// create function to render plain template
func RenderPlainTemplate(temp string, data interface{}) (string, error) {
	// create path to template
	pathToTemplate := fmt.Sprintf("email-templates/%s.plain.tmpl", temp)

	// create template file
	tc, err := template.New("email-template").ParseFS(tempFS, pathToTemplate)

	// check for an error
	if err != nil {
		log.Println("error when creating template html : ", err)
		return "", err
	}

	// create byte to hold data from template\
	var holdData bytes.Buffer

	// execute template
	err = tc.ExecuteTemplate(&holdData, "body", data)

	// cehck for an error
	if err != nil {
		log.Println("error when creating template html with data : ", err)
		return "", err
	}

	// convert byte into string
	result := holdData.String()

	// if okay
	return result, nil
}

// create function to create encyrption
func CreateEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionTLS
	}
}
