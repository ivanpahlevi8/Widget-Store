package main

import (
	"flag"
	"fmt"
	"log"
	"myApp/internal/driver"
	"myApp/internal/models"
	"net/http"
	"os"
	"time"
)

// create application version
const ApplicationVersion = "1.0.0"

// create configuration for application running
type config struct {
	port int
	env  string
	db   struct {
		dns string
	}
	stripe struct {
		key    string
		secret string
	}
	mail        Mail
	secreteKey  string
	frontEndApi string
}

// create application type
/**
application struct digunakna sebagai support applikasi
*/
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	db       models.DbModel
}

// create mail object
type Mail struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	From        string
	FromAddress string
}

func main() {
	// create configuration object
	var config config

	// set configuration member from executable variables
	// this variable are getting from shell script

	// get variable from shell script first
	getPort := flag.Int("port", 4001, "port listenikng for stripes feature")
	getEnv := flag.String("env", "development", "environment condition either production or development")
	getDsn := flag.String("DSN", "root:03052001ivan@tcp(localhost:3305)/widgets?parseTime=true&tls=false", "DSN")
	getStripeKey := "pk_test_51NqzecHIAOtdaeBWDPt9lJOVh42ik6CYub4enGhe5PSROf76bgshIxPDQF3qNXq4W5K9vdHwVg9oW1v4XRpL6wuv00bB4NV5p6"
	getStripeSecret := "sk_test_51NqzecHIAOtdaeBWIxiiQtMi8SC2XgFhRB52he9rlYKxZsf5pg0Wnj5skUhM31UHpYwrbwCEZe3K3yvfUGzrwEVO00gxMqXaDJ"

	// get variable from shell for email
	getHostEmail := flag.String("hostEmail", "sandbox.smtp.mailtrap.io", "host_email")
	getPortEmail := flag.Int("portEmail", 465, "email port")
	getUsernameEmail := flag.String("usernameEmail", "ae07750449f0c6", "email username")
	getPasswordEmail := flag.String("passwordEmail", "3e5baf76732a79", "email password")
	getEncryptionEmail := flag.String("encryptionEmail", "none", "email encryption")
	getFromEmail := flag.String("fromEmail", "test", "from email")
	getFromAddressEmail := flag.String("fromAddressEmail", "test@test.com", "from email address")

	// get variable from shell for goalone for user resdeting password
	getSecreteKey := flag.String("secreteKey", "bRWmrwNUTqNUuzckjxsFlHZjxHkjrzKP", "secrete key reset pass")
	getFrontEnd := flag.String("frontEnd", "http://localhost:4000", "front end api")

	// create mail object
	mail := Mail{
		Host:        *getHostEmail,
		Port:        *getPortEmail,
		Username:    *getUsernameEmail,
		Password:    *getPasswordEmail,
		Encryption:  *getEncryptionEmail,
		From:        *getFromEmail,
		FromAddress: *getFromAddressEmail,
	}

	// set variable to config app
	config.port = *getPort
	config.env = *getEnv
	config.stripe.key = getStripeKey
	config.stripe.secret = getStripeSecret
	config.db.dns = *getDsn
	config.mail = mail
	config.secreteKey = *getSecreteKey
	config.frontEndApi = *getFrontEnd

	// create connection
	conn, err := driver.InitConnection(config.db.dns)
	// check for an error
	if err != nil {
		log.Println("error starting database")
	}
	defer conn.Close()

	// create info log
	// info log akan ditampilkan pada os.Stdout atau pada terminal
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create error log
	// error log akan ditampilkan pada terminal
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create application object
	var myApp = application{
		config:   config,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  ApplicationVersion,
		db:       models.DbModel{DbConn: conn},
	}

	err = myApp.serve()

	// check for an error
	if err != nil {
		myApp.errorLog.Printf("error in listening to server : %s\n", err.Error())
	}
}

// create application function to start http connection
func (app *application) serve() error {
	// create http server object
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	// create log message
	app.infoLog.Printf("application starting in mode : %s, and listening to port : %d\n", app.config.env, app.config.port)

	// listen to server
	err := server.ListenAndServe()

	return err
}
