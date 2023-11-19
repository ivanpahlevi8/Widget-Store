package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"myApp/internal/cards"
	"myApp/internal/encryption"
	"myApp/internal/models"
	"myApp/internal/urlsigner"
	"net/http"
	"strconv"
	"time"
)

// create handler function for testing
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "home", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// create handler function for testing
func (app *Application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "terminal", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

/**
Create Object To Hold Transaction Data
ini digunakan untuk menyimpan data transaksi
data trnasaksi ini disimpan karena terdapat dua cara unutk mencharge kartu kredit
yiaut melalui virtual terminal dan memlalui buy once
sehingga untuk meminimalisir banyaknya code dalam program dapat dijadikan satu
untuk mengambil data transaki yang sama
*/
type TxnData struct {
	FirstName       string
	LastName        string
	CardHolderName  string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	BankReturnCode  string
	ExpiredMonth    int
	ExpiredYear     int
}

// create function to get TxnData from form
func (app *Application) GetTxnData(r *http.Request) (TxnData, error) {
	// create txn data
	txnData := TxnData{}
	// parse form from input body
	err := r.ParseForm()

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when parsing form from body request : %s\n", err.Error())
		return txnData, err
	}

	// get data from form
	getFirstName := r.Form.Get("first_name")
	getLastName := r.Form.Get("last_name")
	getCardholderName := r.Form.Get("cardholder_name")
	getCardHolderEmail := r.Form.Get("cardholder_email")
	getPaymentIntent := r.Form.Get("payment_intent")
	getPaymentMethod := r.Form.Get("payment_method")
	getPaymentAmount := r.Form.Get("payment_amount")
	getPaymentCurrency := r.Form.Get("payment_currency")

	// create card object
	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	// get payment method from card
	pm, err := card.GetCardPaymentMethod(getPaymentMethod)
	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// get paymenyt intent from card
	pi, err := card.GetCardPaymentIntent(getPaymentIntent)
	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// get last four from payment method
	getLastFour := pm.Card.Last4

	// get expired month from payment method
	getExpiredMonth := pm.Card.ExpMonth

	// get expired year from payment method
	getExpiredYear := pm.Card.ExpYear

	// get last four from bank id from payment intent
	getBankCode := pi.Charges.Data[0].ID

	// covert amount to int
	trxAmount, _ := strconv.Atoi(getPaymentAmount)
	trxExpMonth := int(getExpiredMonth)
	trxExpYear := int(getExpiredYear)

	txnData = TxnData{
		FirstName:       getFirstName,
		LastName:        getLastName,
		CardHolderName:  getCardholderName,
		Email:           getCardHolderEmail,
		PaymentIntentID: getPaymentIntent,
		PaymentMethodID: getPaymentMethod,
		PaymentAmount:   trxAmount,
		PaymentCurrency: getPaymentCurrency,
		LastFour:        getLastFour,
		BankReturnCode:  getBankCode,
		ExpiredMonth:    trxExpMonth,
		ExpiredYear:     trxExpYear,
	}

	return txnData, nil

}

// create object for request to microservice
type Invoice struct {
	ID          int       `json:"id"`
	Quantity    int       `json:"quantity"`
	Amount      int       `json:"amount"`
	ProductName string    `json:"product_name"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"-"`
}

// create handler to processing successe payment charge
func (app *Application) SuccededChargePayment(w http.ResponseWriter, r *http.Request) {
	// parse form from input body
	err := r.ParseForm()

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when parsing form from body request : %s\n", err.Error())
		return
	}

	// get transaction data
	trxData, err := app.GetTxnData(r)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when getting transaction data : %s\n", err.Error())
		return
	}

	// get data from form
	getWidgetId := r.Form.Get("product_id")

	/** create customer to database */
	customerId, err := app.SaveCutomer(trxData.FirstName, trxData.LastName, trxData.Email)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	/** create transaction */
	// create transaction object
	transactionObj := models.Transaction{
		Amount:              trxData.PaymentAmount,
		Currency:            trxData.PaymentCurrency,
		LastFour:            trxData.LastFour,
		BankReturnCode:      trxData.BankReturnCode,
		TransactionStatusId: 2,
		ExpiredMonth:        trxData.ExpiredMonth,
		ExpiredYear:         trxData.ExpiredYear,
		PaymentIntent:       trxData.PaymentIntentID,
		PaymentMethod:       trxData.PaymentMethodID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// add transaction to database
	trxId, err := app.SaveTransaction(transactionObj)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	/** create order */
	// convert widget id to int
	widgetId, _ := strconv.Atoi(getWidgetId)

	// create order object
	orderObj := models.Order{
		WidgetID:      widgetId,
		TransactionID: trxId,
		StatusID:      1,
		CustomersID:   customerId,
		Quantity:      1,
		Amount:        trxData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// add order obj to database
	orderId, err := app.SaveOrder(orderObj)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Printf("Getting Order Id : %d\n", orderId)

	/**
	Create logic to handle creating request to invoice microservice to creating
	invoice based on order data
	*/
	// create invoice object
	invoiceObj := Invoice{
		ID:          orderId,
		Quantity:    1,
		Amount:      orderObj.Amount,
		ProductName: "Widget with RGB Color",
		FirstName:   trxData.FirstName,
		LastName:    trxData.LastName,
		Email:       trxData.Email,
		Message:     "Success payment",
		CreatedAt:   time.Now(),
	}

	// do request
	err = app.requestInvoiceMicro(invoiceObj)

	// check fro an error
	if err != nil {
		log.Println("error when requesting for invoice")
		return
	}

	// put data interface to session
	app.Session.Put(r.Context(), "receipt", trxData)

	// redirect user
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

// create function to create request to invoice microservice
func (app *Application) requestInvoiceMicro(inv Invoice) error {
	// cretae url for request
	url := "http://localhost:5000/invoice/createandsendinvoice"

	// create json from object
	jsonObj, err := json.MarshalIndent(
		inv,
		"",
		"\t",
	)

	// check for an error
	if err != nil {
		log.Println("error when converting object into json")
		return err
	}

	// create request object
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(jsonObj),
	)

	// check for an error
	if err != nil {
		log.Println("error when creating request")
		return err
	}

	// set request header as json
	req.Header.Set("Content-Type", "application/json")

	// create client to do request
	client := http.Client{}

	// do request from client
	resp, err := client.Do(req)

	// check for an error
	if err != nil {
		log.Println("error when client do request")
		return err
	}

	defer resp.Body.Close()

	// print response
	log.Println(resp)

	return nil
}

// create handler to processing successe payment charge for virtual terminal
func (app *Application) VirtualTerminalSuccededChargePayment(w http.ResponseWriter, r *http.Request) {
	// get transaction data
	trxData, err := app.GetTxnData(r)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when getting transaction data : %s\n", err.Error())
		return
	}

	/** create transaction */
	// create transaction object
	transactionObj := models.Transaction{
		Amount:              trxData.PaymentAmount,
		Currency:            trxData.PaymentCurrency,
		LastFour:            trxData.LastFour,
		BankReturnCode:      trxData.BankReturnCode,
		TransactionStatusId: 2,
		ExpiredMonth:        trxData.ExpiredMonth,
		ExpiredYear:         trxData.ExpiredYear,
		PaymentIntent:       trxData.PaymentIntentID,
		PaymentMethod:       trxData.PaymentMethodID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// add transaction to database
	_, err = app.SaveTransaction(transactionObj)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// put data interface to session
	app.Session.Put(r.Context(), "receipt", trxData)

	// redirect user
	http.Redirect(w, r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

// create handler function to show receipt to user
func (app *Application) ShowReceipt(w http.ResponseWriter, r *http.Request) {
	// get data from session
	trxData := app.Session.Get(r.Context(), "receipt").(TxnData)

	// create data interface
	dataInterface := make(map[string]interface{})

	// put trx data
	dataInterface["trx"] = trxData

	// create data template
	templateData := TemplateData{
		DataMap: dataInterface,
	}

	// render template
	err := app.renderTemplate(w, r, "succededd", &templateData)

	// check fo an error
	if err != nil {
		app.errorLog.Printf("error when rendering template : %s\n", err.Error())
		return
	}
}

// create handler function to show receipt to user
func (app *Application) VirtualTerminalShowReceipt(w http.ResponseWriter, r *http.Request) {
	// get data from session
	trxData := app.Session.Get(r.Context(), "receipt").(TxnData)

	// create data interface
	dataInterface := make(map[string]interface{})

	// put trx data
	dataInterface["trx"] = trxData

	// create data template
	templateData := TemplateData{
		DataMap: dataInterface,
	}

	// render template
	err := app.renderTemplate(w, r, "virtual-terminal-succededd", &templateData)

	// check fo an error
	if err != nil {
		app.errorLog.Printf("error when rendering template : %s\n", err.Error())
		return
	}
}

// create function to access database customer
func (app *Application) SaveCutomer(firstName string, lastName string, email string) (int, error) {
	// create customer object
	customerObj := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// add customer obj to database
	customerId, err := app.db.InsertCustomer(customerObj)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when adding customer to datavbase : %s\n", err)
		return -1, err
	}

	// if success
	return customerId, nil
}

// create function to access database transaction
func (app *Application) SaveTransaction(trx models.Transaction) (int, error) {
	// add transaction to database
	trxId, err := app.db.InsertTransaction(trx)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when adding transaction to datavbase : %s\n", err)
		return -1, err
	}

	// if success
	return trxId, nil
}

// create function to access database order
func (app *Application) SaveOrder(order models.Order) (int, error) {
	// add transaction to database
	orderId, err := app.db.InsertOrder(order)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when adding order to datavbase : %s\n", err)
		return -1, err
	}

	// if success
	return orderId, nil
}

// create handler to show widget page
func (a *Application) OncePage(w http.ResponseWriter, r *http.Request) {
	// set header
	w.WriteHeader(http.StatusOK)

	// parse id from url
	queryUrl := r.URL.Query()

	// get id from url
	getId := queryUrl.Get("id")

	// convert id from string to integer
	idWidget, err := strconv.Atoi(getId)
	log.Println("id widget : ", idWidget)

	// check for an error
	if err != nil {
		a.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create widget object
	widgets, err := a.db.GetWidgetById(idWidget)

	// check for an error
	if err != nil {
		a.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create data map
	dataMap := make(map[string]interface{})

	//put model to data map
	dataMap["widgets"] = widgets

	// create template data
	td := TemplateData{
		DataMap: dataMap,
	}

	// render template
	err = a.renderTemplate(w, r, "buy-one", &td, "stripe-js")

	// check for an erro
	if err != nil {
		a.errorLog.Println("error when rendering buy one template")
	}
}

// create handler to show subsription page
func (app *Application) BronzeSubscription(w http.ResponseWriter, r *http.Request) {
	// create datamap to hold bronze id
	intMap := make(map[string]int)

	// create interface map
	interfaceMap := make(map[string]interface{})

	// put bronze id to int map
	intMap["bronze_id"] = 1

	// get widget by id
	getWidget, err := app.db.GetWidgetById(2)

	// put widget in interface map
	interfaceMap["widget"] = getWidget

	log.Println("widget price : ", getWidget.Price)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when get widget by id : ", err)
		return
	}

	// create data map
	dataMap := TemplateData{
		IntMap:  intMap,
		DataMap: interfaceMap,
	}

	// render
	err = app.renderTemplate(w, r, "bronze-page", &dataMap)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when rendering bronze page : ", err)
		return
	}
}

// create bronze receipt
func (app *Application) ShowBronzeReceipt(w http.ResponseWriter, r *http.Request) {
	// render
	err := app.renderTemplate(w, r, "bronze-subs-receipt", &TemplateData{})

	// check for an error
	if err != nil {
		app.errorLog.Println("error when rendering bronze page : ", err)
		return
	}
}

// create function to show login page
func (app *Application) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	// render
	err := app.renderTemplate(w, r, "login", &TemplateData{})

	// check for an error
	if err != nil {
		app.errorLog.Println("error when rendering bronze page : ", err)
		return
	}
}

// create function to process login
func (app *Application) ProcessLogin(w http.ResponseWriter, r *http.Request) {
	// renew token
	app.Session.RenewToken(r.Context())

	log.Println("process login passed")

	// parse form
	err := r.ParseForm()

	// check for an error
	if err != nil {
		log.Println("error happend : ", err)
		app.errorLog.Println("error when parsing form : ", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// get value from form
	getEmailInput := r.Form.Get("email")
	getPasswordInput := r.Form.Get("password")

	// authenticated user
	getId, err := app.db.ValidateUser(getEmailInput, getPasswordInput)

	// check for an error
	if err != nil {
		log.Println(err)
		app.errorLog.Println(err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// if successs, save id to session
	app.Session.Put(r.Context(), "user_id", getId)

	// redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// crtete function to logout
func (app *Application) LogoutProcess(w http.ResponseWriter, r *http.Request) {
	// clear session
	app.Session.Destroy(r.Context())

	// renewq token
	app.Session.RenewToken(r.Context())

	// redirect user to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// create fucntion to show rest password page
func (app *Application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// render template
	err := app.renderTemplate(
		w,
		r,
		"reset-password",
		&TemplateData{},
	)

	// checkl for an error
	if err != nil {
		app.errorLog.Println("error when loading reset password page : ", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// create hanbdler to show reset password page
func (app *Application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {
	// get url request
	urlParamReq := r.RequestURI

	// cretae test url
	testUrl := fmt.Sprintf("%s%s", app.config.frontEndApi, urlParamReq)
	log.Println(testUrl)

	// check tokens
	urlCheck := urlsigner.Signer{
		Secret: []byte(app.config.secreteKey),
	}

	// cehck token
	isOk, err := urlCheck.ValidateToken(testUrl)

	if err != nil {
		log.Println(err)
		w.Write([]byte("invalid token"))
		return
	}

	if !isOk {
		w.Write([]byte("invalid token"))
		return
	}

	// validate token expiry for 60 minute
	isExpiry := urlCheck.ValidateDuration(testUrl, 60)

	if !isExpiry {
		w.Write([]byte("invalid token, token is already expired"))
		return
	}

	// query url to get email data
	query := r.URL.Query()

	// get email from url
	getEmail := query.Get("email")

	// create encryptionb object to encrypt email
	encrytption := encryption.Encryption{
		Key: []byte(app.config.secreteKey),
	}

	// create encrypted email
	encryptedEmail, err := encrytption.EncryptionText(getEmail)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when encrypted email: ", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// create string data map
	strDataMap := make(map[string]string)

	// assign data to string map
	strDataMap["email"] = encryptedEmail

	// crete data template to pass to template
	dataTemplate := TemplateData{
		StringMap: strDataMap,
	}

	// render template
	err = app.renderTemplate(w, r, "show-reset-password", &dataTemplate)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when loading reset password page to show page : ", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// create fucntion to show all sales page
func (app *Application) ShowAllSales(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "all-sales", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// create fucntion to show all sales page
func (app *Application) ShowAllSubs(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "all-subs", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// cretae fucntion to show single sale
func (app *Application) ShowSale(w http.ResponseWriter, r *http.Request) {
	// cretae template string data
	dataString := make(map[string]string)

	// assign data to datastring
	dataString["title"] = "Sale"
	dataString["head"] = "Single Sale"
	dataString["back_link"] = "/admin/all-sales"

	// add data template to shown wether it is a refunc or subs cancle
	dataString["btn-info"] = "Refund Charge"
	dataString["url-info"] = "http://localhost:4001/api/admin/refund"
	dataString["dialog-info1"] = "You won't be able to undo this refund!"
	dataString["dialog-confirm-button"] = "Yes, refund it!"
	dataString["success-message"] = "Success Refunding Charged Payment!!!"
	dataString["badge-info"] = "Refunded"

	// create template data
	td := TemplateData{
		StringMap: dataString,
	}

	err := app.renderTemplate(w, r, "sale", &td)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// cretae fucntion to show single sale
func (app *Application) ShowSubs(w http.ResponseWriter, r *http.Request) {
	// cretae template string data
	dataString := make(map[string]string)

	// assign data to datastring
	dataString["title"] = "Subscription"
	dataString["head"] = "Single Subscription"
	dataString["back_link"] = "/admin/all-sub"

	// add data template to shown wether it is a refunc or subs cancle
	dataString["btn-info"] = "Cancel Subscription"
	dataString["url-info"] = "http://localhost:4001/api/admin/cancel"
	dataString["dialog-info1"] = "You won't be able to undo this cancel subscription!"
	dataString["dialog-confirm-button"] = "Yes, cancel it!"
	dataString["success-message"] = "Success Canceling Subscription Payment!!!"
	dataString["badge-info"] = "Canceled"

	// create template data
	td := TemplateData{
		StringMap: dataString,
	}

	err := app.renderTemplate(w, r, "sale", &td)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// create function to show all users
func (app *Application) ShowAllUsers(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "all-users", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}

// create function to show single user
func (app *Application) ShowUser(w http.ResponseWriter, r *http.Request) {
	err := app.renderTemplate(w, r, "user", nil)

	if err != nil {
		app.errorLog.Println(err)
	}
}
