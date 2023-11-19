package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"myApp/internal/driver"
	"myApp/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
)

// cretae session object as public
var sessionManager *scs.SessionManager

// create application version
const ApplicationVersion = "1.0.0"

// create configuration for application running
type Config struct {
	port int
	env  string
	api  string
	db   struct {
		dns string
	}
	stripe struct {
		key    string
		secret string
	}
	secreteKey  string
	frontEndApi string
}

// create application type
/**
application struct digunakna sebagai support applikasi
*/
type Application struct {
	config   Config
	infoLog  *log.Logger
	errorLog *log.Logger
	tc       map[string]*template.Template
	version  string
	db       models.DbModel
	Session  *scs.SessionManager
}

// create main function
func main() {
	// register object to session
	/**
	register dilakukan agar session dapat menerima data interface dari receipt
	*/
	gob.Register(map[string]interface{}{})

	// register transaction data
	gob.Register(TxnData{})

	// create configuration object
	var config Config

	// set configuration member from executable variables
	// this variable are getting from shell script

	// get variable from shell script first
	getPort := flag.Int("port", 4000, "port listenikng for stripes feature")
	getEnv := flag.String("env", "development", "environment condition either production or development")
	getApi := flag.String("api", "localhost:4001", "Url to access api")
	getDsn := flag.String("DSN", "root:03052001ivan@tcp(localhost:3305)/widgets?parseTime=true&tls=false", "DSN")
	getStripeKey := "pk_test_51NqzecHIAOtdaeBWDPt9lJOVh42ik6CYub4enGhe5PSROf76bgshIxPDQF3qNXq4W5K9vdHwVg9oW1v4XRpL6wuv00bB4NV5p6"
	getStripeSecret := "sk_test_51NqzecHIAOtdaeBWIxiiQtMi8SC2XgFhRB52he9rlYKxZsf5pg0Wnj5skUhM31UHpYwrbwCEZe3K3yvfUGzrwEVO00gxMqXaDJ"

	// get variable from shell for goalone for user resdeting password
	getSecreteKey := flag.String("secreteKey", "bRWmrwNUTqNUuzckjxsFlHZjxHkjrzKP", "secrete key reset pass")
	getFrontEnd := flag.String("frontEnd", "http://localhost:4000", "front end api")

	// set variable to config app
	config.port = *getPort
	config.env = *getEnv
	config.api = *getApi
	config.stripe.key = getStripeKey
	config.stripe.secret = getStripeSecret
	config.db.dns = *getDsn
	config.secreteKey = *getSecreteKey
	config.frontEndApi = *getFrontEnd

	// create connection
	conn, err := driver.InitConnection(config.db.dns)
	// check for an error
	if err != nil {
		log.Println("error starting database")
	}
	defer conn.Close()

	// create session
	sessionManager = scs.New()
	sessionManager.Store = mysqlstore.New(conn)

	// set session lifetime
	sessionManager.Lifetime = 24 * time.Hour

	// create info log
	// info log akan ditampilkan pada os.Stdout atau pada terminal
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create error log
	// error log akan ditampilkan pada terminal
	errorLog := log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create template map cache
	// diogunakan untuk menyimpan template dari html
	templateCache := make(map[string]*template.Template)

	// create application object
	var myApp = Application{
		config:   config,
		infoLog:  infoLog,
		errorLog: errorLog,
		tc:       templateCache,
		version:  ApplicationVersion,
		db:       models.DbModel{DbConn: conn},
		Session:  sessionManager,
	}

	// listen forever from websocket
	go myApp.ListenToWSChannel()

	err = myApp.serve()

	// check for an error
	if err != nil {
		myApp.errorLog.Printf("error in listening to server : %s\n", err.Error())
	}
}

// create application function to start http connection
func (app *Application) serve() error {
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
