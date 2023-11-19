package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"myApp/internal/cards"
	"myApp/internal/encryption"
	"myApp/internal/models"
	"myApp/internal/urlsigner"
	"myApp/internal/validators"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	"golang.org/x/crypto/bcrypt"
)

// create stripe payload
/**
stripe payload digunakan untuk menampung data mengenai pembayaran seperti jumlah dan currency
yang digunakan
*/

type StripePayload struct {
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Plan          string `json:"plan"`
	PaymentMethod string `json:"payment_method"`
	LastFour      string `json:"last_four"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"exp_month"`
	ExpiryYear    int    `json:"exp_year"`
	ProductId     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
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
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	CardHolderName  string `json:"cardholder_name"`
	Email           string `json:"email"`
	PaymentIntentID string `json:"payment_intent"`
	PaymentMethodID string `json:"payment_method"`
	PaymentAmount   int    `json:"payment_amount"`
	PaymentCurrency string `json:"payment_currency"`
	LastFour        string `json:"last_four"`
	BankReturnCode  string `json:"return_code"`
	ExpiredMonth    int    `json:"expired_month"`
	ExpiredYear     int    `json:"expired_year"`
}

/**
create paymen intent handler to show payment intent as json response
*/
func (a *application) PaymentIntentHandler(w http.ResponseWriter, r *http.Request) {
	// set header as json response
	w.Header().Set("Content-Type", "application/json")

	// create variable to hold payload from body request
	var payloadBody = StripePayload{}

	// get payloiad from body
	/*
		payload dari body merupakan suatu json, jadi harus dilakukan proses decoding dari json
		menjadi suatu obejct struct
	*/
	err := json.NewDecoder(r.Body).Decode(&payloadBody)

	// check for an error
	if err != nil {
		a.errorLog.Printf("error when decoding json from body payload : %s\n", err.Error())
		return
	}

	// get value amount from payload as integer

	getAmount, err := strconv.ParseFloat(payloadBody.Amount, 32)

	// check for an error
	if err != nil {
		a.errorLog.Printf("error when decoding json from body payload : %s\n", err.Error())
		return
	}

	// get value currency from payload
	getCurrency := payloadBody.Currency

	// create card object
	cardObj := cards.Card{
		Key:      a.config.stripe.key,
		Secret:   a.config.stripe.secret,
		Currency: getCurrency,
	}

	// charge card baed on credentials
	payloadIntent, msg, err := cardObj.Charge(getCurrency, int(getAmount))

	// create var to state okay or not
	okay := true

	// check for an error
	if err != nil {
		okay = false
	}

	// check okay
	if okay {
		// create json object from payload intent
		jsonReturn, err := json.MarshalIndent(payloadIntent, "", "    ")

		// check for an error
		if err != nil {
			a.errorLog.Printf("error when decoding json from body payload : %s\n", err.Error())
			return
		}

		// write to web
		w.Write(jsonReturn)
	} else {
		log.Println("error processing : ", payloadIntent)
		// cretae json response for testing
		response := JsonResponse{
			OK:      false,
			Message: "Error JSON Response When Charge Credit Card",
			Content: msg,
			Id:      uuid.New().String(),
		}

		// marshal object into json format
		responseJson, err := json.MarshalIndent(response, "", "      ")

		// check for an error
		if err != nil {
			a.errorLog.Println("errro when converting object into json response")
			return
		}

		// show json in web
		w.Write(responseJson)
	}
}

// create handler to test getting
func (app *application) ShowWdigetById(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// parse id from url
	queryUrl := r.URL.Query()

	// get id from url
	getId := queryUrl.Get("id")

	// convert id from string to integer
	idWidget, err := strconv.Atoi(getId)
	log.Println("id widget : ", idWidget)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get widget
	getWidget, err := app.db.GetWidgetById(idWidget)

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal json
	jsonObj, err := json.MarshalIndent(getWidget, "", "\t")

	// check for an error
	if err != nil {
		app.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set header
	w.WriteHeader(http.StatusOK)
	w.Write(jsonObj)
}

func (app *application) AddCustomerAndSubcsription(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// cretae stripe payload to hold value from body request
	var paylodStripe StripePayload

	// get json value from body request
	err := json.NewDecoder(r.Body).Decode(&paylodStripe)

	// print value
	log.Printf("%s | %s | %s | %s | %s | %s | %d |%d\n", paylodStripe.Email, paylodStripe.LastFour, paylodStripe.PaymentMethod, paylodStripe.Plan, paylodStripe.Amount, paylodStripe.ProductId, paylodStripe.ExpiryMonth, paylodStripe.ExpiryYear)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when parsing body from request : ", err)
		return
	}

	// valiate payloadStripe
	// create validate object
	validate := validators.New()

	// check for data in payload
	log.Println("More than 3 words : ", len(paylodStripe.FirstName) > 3)
	validate.Check(len(paylodStripe.FirstName) > 3, "first_name", "Name must contains at least 4 character")
	log.Println("check validate : ", validate.Error["first_name"])

	// validate
	if !validate.Validate() {
		app.fieldValidation(w, r, validate.Error)
		return
	}

	// create card object
	card := cards.Card{
		Key:      app.config.stripe.key,
		Secret:   app.config.stripe.secret,
		Currency: paylodStripe.Currency,
	}

	// cretae variable to check error existing
	ok := true

	// get customer from card subscription
	getCustomer, msg, err := card.CreateCustomer(paylodStripe.PaymentMethod, paylodStripe.Email)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when creating user for subscription : ", err, " Msg : ", msg)
		ok = false
		return
	}

	var getSubscription *stripe.Subscription

	if ok {
		// create subscription based on customer
		getSubscription, err = card.AddNewSubscription(
			getCustomer,
			paylodStripe.LastFour,
			paylodStripe.PaymentMethod,
			paylodStripe.Email,
			"",
			paylodStripe.Plan)
		/**
		subscription akan memiliki nilai id, dimana nilai id tersebut akan dsimpan sebagai
		PaymentIntent pada transaction object untuk digunakan sebagai Id payment intent
		pada method unsubscribe yang akan dilakukan terhadap order tersebut!!!
		*/

		// check for an error
		if err != nil {
			app.errorLog.Println("errro when creating subscription based on customer : ", err)
			ok = false
			return
		}
	}

	app.infoLog.Println("getting subscription id : ", getSubscription.ID)

	// create response json
	var response JsonResponse

	// check if ok or not ok
	if !ok {
		// creae response
		response = JsonResponse{
			OK:      false,
			Message: "Error when creating subscription customers",
			Content: "This content back from backend",
			Id:      uuid.New().String(),
		}
		// convert response to json
		respJson, err := json.MarshalIndent(response, "", "\t")

		// check for an error
		if err != nil {
			app.errorLog.Println("error when marshalling obejct to json : ", err)
			return
		}

		// write resposne to json
		w.Write(respJson)
	} else {
		// if success and there is no error
		/**
		jika sukses, maka buat customer, transaction, and order yang akan dsimpan pada database
		*/

		// create customer
		createCustomer := models.Customer{
			FirstName: paylodStripe.FirstName,
			LastName:  paylodStripe.LastName,
			Email:     paylodStripe.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// save customer to database
		customerId, err := app.SaveCutomer(createCustomer)

		// check for an error
		if err != nil {
			app.errorLog.Println("error when save customer to database")
			return
		}

		// convert amount to int
		getAmount, err := strconv.Atoi(paylodStripe.Amount)
		if err != nil {
			app.errorLog.Println("error when converting amount to int")
			return
		}

		// create transaction object
		txn := models.Transaction{
			Amount:              getAmount,
			Currency:            paylodStripe.Currency,
			LastFour:            paylodStripe.LastFour,
			TransactionStatusId: 2,
			ExpiredMonth:        paylodStripe.ExpiryMonth,
			ExpiredYear:         paylodStripe.ExpiryYear,
			PaymentIntent:       getSubscription.ID, // digunakan untuk memanggil method untuk mengunsubscribe subscription
			PaymentMethod:       paylodStripe.PaymentMethod,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// save transaction to database
		txnId, err := app.SaveTransaction(txn)

		// check for an error
		if err != nil {
			app.errorLog.Println("error when save transaction to database")
			return
		}

		// convert product id to int
		getProductId, err := strconv.Atoi(paylodStripe.ProductId)
		if err != nil {
			app.errorLog.Println("error when converting product id to int")
			return
		}

		// create order object
		order := models.Order{
			WidgetID:      getProductId,
			TransactionID: txnId,
			StatusID:      1,
			CustomersID:   customerId,
			Quantity:      1,
			Amount:        getAmount,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// save order
		orderId, err := app.SaveOrder(order)

		// check for an error
		if err != nil {
			app.errorLog.Println("error when save order to database")
			return
		}

		// creae response
		response = JsonResponse{
			OK:      true,
			Message: "response back from back end, 101",
			Content: fmt.Sprintf("getting order id from back end : %d\n", orderId),
			Id:      uuid.New().String(),
		}

		// convert response to json
		respJson, err := json.MarshalIndent(response, "", "\t")

		// check for an error
		if err != nil {
			app.errorLog.Println("error when marshalling obejct to json : ", err)
			return
		}

		// write resposne to json
		w.Write(respJson)
	}
}

// create function to access database customer
func (app *application) SaveCutomer(customer models.Customer) (int, error) {
	// add customer obj to database
	customerId, err := app.db.InsertCustomer(customer)

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when adding customer to datavbase : %s\n", err)
		return -1, err
	}

	// if success
	return customerId, nil
}

// create function to access database transaction
func (app *application) SaveTransaction(trx models.Transaction) (int, error) {
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
func (app *application) SaveOrder(order models.Order) (int, error) {
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

// cretae function to authentication user
func (app *application) AuthenticationUser(w http.ResponseWriter, r *http.Request) {
	// set header request
	w.Header().Set("Content-Type", "application/json")

	// create payload object from request
	var requestPayload UserAuthPayload

	// create response payload
	var responsePayload JsonResponse

	// read request from body
	err := json.NewDecoder(r.Body).Decode(&requestPayload)

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when reading request body : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get user by email from payload
	emailCasting := strings.ToLower(requestPayload.Email)
	getUser, err := app.db.GetUserByEmail(emailCasting)

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when get user by id in auth : ", err)
		app.InvalidAuthUser(w, requestPayload.Email)
		return
	}

	// get password from user and compare with user request password
	passFromUserDatabase := getUser.Password

	// compare error
	compare, err := app.MatchesPassword(passFromUserDatabase, requestPayload.Password)

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when mathcing password user and input : ", err)
		app.InvalidAuthUser(w, requestPayload.Email)
		return
	}

	// check if mathcing valid
	if !compare {
		// if user input wrong password
		log.Println("error invalid passowrd input : ", err)
		app.InvalidAuthUser(w, requestPayload.Email)
		return
	}

	// create user token as prove that user already authenticated
	userToken, err := models.GenerateToke(int64(getUser.ID), 24*time.Hour, "authentication")

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when creating user token : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// add token to datavase
	getId, err := app.db.AddTokenToDatabase(userToken, getUser)

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when input token to database : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// asisgn response
	responsePayload = JsonResponse{
		OK:      false,
		Message: "user success to authentication",
		Content: fmt.Sprintf("user with email : %s, is authenticated with token id : %d", requestPayload.Email, getId),
		Id:      uuid.New().String(),
		Token:   userToken,
	}

	err = app.WriteJsonObject(w, responsePayload, http.StatusAccepted)

	// check for an error
	if err != nil {
		// if error happen
		log.Println("error when reading request body : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}
}

// create function to validate token
/**
validate token akan memvalidasi token. Jika token tervalidasi akan dilakuakan penngmbilan data user berdasarkan
nilai token yang didapatkan dari header. Kemudia return user akan direturn sebagai model authenticated
*/
func (app *application) ValidateToken(r *http.Request) (*models.User, error) {
	// get authorization parts from header
	getHeader := r.Header.Get("Authorization")

	// check header
	if getHeader == "" {
		// if there is no error
		log.Println("error no header authorization in request, please login!!!")
		return nil, errors.New("error no header authorization in request, please login")
	}

	// split authorization by space
	headerData := strings.Split(getHeader, " ")

	// check error data
	if len(headerData) != 2 || headerData[0] != "Bearer" {
		// if data in header data is not appropriate
		log.Println("error data in authorization header are not correct, please re login!!!")
		return nil, errors.New("error data in authorization header are not correct, please re login")
	}

	// get token from header data in second argument
	getTokens := headerData[1]

	// check token len
	if len(getTokens) != 26 {
		log.Println("token lenght is not valid")
		return nil, errors.New("token lenght is not valid")
	}

	// get user by tokens
	getUser, err := app.db.GetUserForTokens(getTokens)

	// check fro an error
	if err != nil {
		log.Println("error when getting user for tokens : ", err)
		return nil, err
	}

	// if success
	return getUser, nil
}

// creatre fucntion to check auth
func (a *application) CheckAuth(w http.ResponseWriter, r *http.Request) {
	// authenticated tokens
	getUser, err := a.ValidateToken(r)

	// check for an error
	if err != nil {
		log.Println("not authenticated user : ", err)
		a.InvalidAuthUser(w, "test@test.com")
		return
	}

	// create response
	var response = JsonResponse{
		OK:      false,
		Message: fmt.Sprintf("user successfull authenticated user : %s", getUser.Email),
	}

	a.WriteJsonObject(w, response, http.StatusAccepted)
}

// create fucntion to processing virtual terminal
func (app *application) ProcessSucceddedPayment(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// create object to hold data from body request
	var txnData TxnData

	// get data from body request
	err := json.NewDecoder(r.Body).Decode(&txnData)

	// check for an error
	if err != nil {
		log.Println("error when parsing body request to object data : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create card object
	card := cards.Card{
		Key:    app.config.stripe.key,
		Secret: app.config.stripe.secret,
	}

	// creare payment method
	pm, err := card.GetCardPaymentMethod(txnData.PaymentMethodID)

	// check for an error
	if err != nil {
		log.Println("error when creating payment method : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create payment intent
	pi, err := card.GetCardPaymentIntent(txnData.PaymentIntentID)

	// check for an error
	if err != nil {
		log.Println("error when creating payment intent : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get last four
	getLastFour := pm.Card.Last4

	// get expired month
	getExpiredMonth := pm.Card.ExpMonth

	// get expired year
	getExpiredYear := pm.Card.ExpYear

	// get bank return code
	bankReturnCode := pi.Charges.Data[0].ID

	// change to integer
	expMonth := int(getExpiredMonth)
	expYear := int(getExpiredYear)

	// assign empty valu in txn data
	txnData.LastFour = getLastFour
	txnData.BankReturnCode = bankReturnCode
	txnData.ExpiredMonth = expMonth
	txnData.ExpiredYear = expYear

	// create transaction model
	transactionObj := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		BankReturnCode:      txnData.BankReturnCode,
		TransactionStatusId: 2,
		ExpiredMonth:        expMonth,
		ExpiredYear:         expYear,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// save to database
	_, err = app.SaveTransaction(transactionObj)

	// check for an error
	if err != nil {
		log.Println("error when saving transaction to database : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// send response to user
	app.WriteJsonObject(w, transactionObj, http.StatusAccepted)
}

// create request payload for reset password
type ResetPassword struct {
	Email string `json:"email"`
}

// create fucntion to handle request reset password
func (app *application) ResetPasswordMailer(w http.ResponseWriter, r *http.Request) {
	// set header as json apps
	w.Header().Set("Content-Type", "application/json")

	// create object to hold value from body
	var resetObj ResetPassword

	// read from body requesty
	err := app.ReadJsonBodyRequest(w, r, &resetObj)

	// check for an error
	if err != nil {
		log.Println("error when reading object from body request")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get email from payload
	getEmail := resetObj.Email

	// check if user exist or not
	_, err = app.db.GetUserByEmail(getEmail)

	// check for an error
	if err != nil {
		log.Println("error when getting user with email : ", err)
		// create response
		responseUserErr := JsonResponse{
			OK:      true,
			Message: "email is not valid",
			Content: fmt.Sprintf("error : %s", err),
			Id:      uuid.New().String(),
		}

		// write response
		app.WriteJsonObject(w, responseUserErr, http.StatusInternalServerError)
		return
	}

	// create email message based on email...

	// create signer
	signer := urlsigner.Signer{
		Secret: []byte(app.config.secreteKey),
	}

	// create url token
	urlWillAdd := fmt.Sprintf("%s/reset-password?email=%s", app.config.frontEndApi, getEmail)
	log.Println(urlWillAdd)
	urlToken := signer.CreateNewToken(urlWillAdd)

	// create email template data
	emailData := EmailTD{
		Link: urlToken,
	}

	// send email
	err = app.SendEmail(
		"test@test.com",
		"test@test.com",
		"change password",
		"password-reset",
		&emailData,
	)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when sending email : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create response
	responseUser := JsonResponse{
		OK:      false,
		Message: "success sending email reset password to user",
		Content: "reset password content",
		Id:      uuid.New().String(),
	}

	// send response
	app.WriteJsonObject(w, responseUser, http.StatusCreated)
}

// create request payload for reset password
type ProcessResetPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// create function to process reset password
func (app *application) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	// set content header
	w.Header().Set("Content-Type", "application/json")

	// create object to read body request
	var payloadObj ProcessResetPassword

	// read json from request body
	err := app.ReadJsonBodyRequest(w, r, &payloadObj)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when read json object from body request : ", err)
		// create response
		responseError := JsonResponse{
			OK:      true,
			Message: "error read json from request body",
			Content: err.Error(),
		}
		// write response
		app.WriteJsonObject(w, responseError, http.StatusInternalServerError)
		return
	}

	// get email and password from payload
	getEmailHash := payloadObj.Email

	// create encryption object to decrypt email
	encryp := encryption.Encryption{
		Key: []byte(app.config.secreteKey),
	}

	// decrypot object
	getEmail, err := encryp.DecodeText(getEmailHash)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when decrypted email : ", err)
		// create response
		responseError := JsonResponse{
			OK:      true,
			Message: "error when decrypoted email",
			Content: err.Error(),
		}
		// write response
		app.WriteJsonObject(w, responseError, http.StatusInternalServerError)
		return
	}

	getPassword := payloadObj.Password

	log.Println("email in payload : ", getEmail)

	// get user by email for reseting password
	getUser, err := app.db.GetUserByEmail(getEmail)

	// check for an error
	if err != nil {
		app.errorLog.Println("error when get user by email from database : ", err)
		// create response
		responseError := JsonResponse{
			OK:      true,
			Message: "error get user by email from datrabase",
			Content: err.Error(),
		}
		// write response
		app.WriteJsonObject(w, responseError, http.StatusInternalServerError)
		return
	}

	// create hash password from bycrypt
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(getPassword), 12)

	// check for an error
	if err != nil {
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// update password in database
	err = app.db.UpdateUserPassword(getUser, string(hashPassword))

	// check for an error
	if err != nil {
		app.errorLog.Println("error when updating user password in database : ", err)
		// create response
		responseError := JsonResponse{
			OK:      true,
			Message: "error when update user passwrod in databvase",
			Content: err.Error(),
		}
		// write response
		app.WriteJsonObject(w, responseError, http.StatusInternalServerError)
		return
	}

	// if okay create response payload
	response := JsonResponse{
		OK:      false,
		Message: "Success updating password",
		Content: fmt.Sprintf("user with email, %s, successfully updating password!", getUser.Email),
		Id:      uuid.New().String(),
	}

	// send response
	app.WriteJsonObject(w, response, http.StatusCreated)
}

// create object for receive data for paginated all sales page and all subs page
type PaginatedRequest struct {
	PageSize    int `json:"page_size"`
	CurrentPage int `json:"current_page"`
}

// create object for sending data for paginated all sales page and all subs page
type PaginatedResponse struct {
	PageSize    int            `json:"page_size"`
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
	AllPage     int            `json:"all_page"`
	Order       []models.Order `json:"all_order"`
}

// create fucntion to get all slaed
func (app *application) GetAllSales(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// create request payload object
	var requestPayload PaginatedRequest

	// read from response
	err := app.ReadJsonBodyRequest(w, r, &requestPayload)

	// check for an error
	if err != nil {
		log.Println("error when reading request for paginated sales page")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get all sales data
	allSales, allData, lastPage, err := app.db.GetAllOrdersPaginated(requestPayload.PageSize, requestPayload.CurrentPage)

	// check error
	if err != nil {
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create response
	responsePayload := PaginatedResponse{
		PageSize:    requestPayload.PageSize,
		CurrentPage: requestPayload.CurrentPage,
		LastPage:    lastPage,
		AllPage:     allData,
		Order:       allSales,
	}

	// write to json
	app.WriteJsonObject(w, responsePayload, http.StatusAccepted)
}

// create function to show all subscription orders
func (app *application) GetAllSubscription(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// create object for request payload
	var requestPayload PaginatedRequest

	// get request from body
	err := app.ReadJsonBodyRequest(w, r, &requestPayload)

	// check for an error
	if err != nil {
		log.Println("error when reading request payload from all subscription")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get all sales data
	allSales, allData, pages, err := app.db.GetAllSubscriptionPaginated(requestPayload.PageSize, requestPayload.CurrentPage)

	// check error
	if err != nil {
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create response
	responsePayload := PaginatedResponse{
		PageSize:    requestPayload.PageSize,
		CurrentPage: requestPayload.CurrentPage,
		LastPage:    pages,
		AllPage:     allData,
		Order:       allSales,
	}

	// write to json
	app.WriteJsonObject(w, responsePayload, http.StatusAccepted)
}

// cretae function to showl single sales based on id
func (app *application) ShowSalesById(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// create query
	query := r.URL.Query()

	// get id from url
	getIdStr := query.Get("id")

	// convert id to integer
	getId, _ := strconv.Atoi(getIdStr)

	// get order by id
	order, err := app.db.GetOrderById(getId)

	// check for an error
	if err != nil {
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// if success
	app.WriteJsonObject(w, order, http.StatusAccepted)
}

// create object to hold value from refund request
type RefundRequestData struct {
	ID            int    `json:"id"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
	PaymentIntent string `json:"payment_intent"`
}

// create function to do refund purchasing
func (app *application) RefundPurchasing(w http.ResponseWriter, r *http.Request) {
	// set header as json application
	w.Header().Set("Content-Type", "application/json")

	// create refund object to hold value from json request
	var refundObj RefundRequestData

	// read json from object
	err := app.ReadJsonBodyRequest(w, r, &refundObj)

	// check for an error
	if err != nil {
		log.Println("error when getting json request in refund purchasing")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create card object to access refund function
	cardObject := cards.Card{
		Key:      app.config.stripe.key,
		Secret:   app.config.stripe.secret,
		Currency: refundObj.Currency,
	}

	// do refund function based on card object
	err = cardObject.RefundPayment(refundObj.PaymentIntent, refundObj.Amount)

	// check for an error
	if err != nil {
		log.Println("error when do refund from payment")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// update status on database
	err = app.db.UpdateOrderStatus(refundObj.ID, 2)

	// check for an error
	if err != nil {
		log.Println("error when updating order status on database")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, errors.New("success to refund charged, but failed to update status in database"))
		return
	}

	// if success creating success response
	responseUser := JsonResponse{
		OK:      false, // means does'nt have any error when requesting refund
		Message: "success refunding",
		Content: fmt.Sprintf("success refund purchasing with id : %d", refundObj.ID),
		Id:      uuid.New().String(),
	}

	// send response
	app.WriteJsonObject(w, &responseUser, http.StatusAccepted)
}

// create object to hold data from cancle subscription request
type CancleSubObject struct {
	ID            int    `json:"id"`
	PaymentIntent string `json:"payment_intent"`
	Currency      string `json:"currency"`
}

// create function to cancle subscription
func (app *application) CancleSubscription(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// creaete object
	var requestObj CancleSubObject

	// read json from request
	err := app.ReadJsonBodyRequest(w, r, &requestObj)

	// check for an error
	if err != nil {
		log.Println("error when reading request from cancle subscription")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	log.Println("subscription id : ", requestObj.PaymentIntent)

	// create card object to access refund function
	cardObject := cards.Card{
		Key:      app.config.stripe.key,
		Secret:   app.config.stripe.secret,
		Currency: requestObj.Currency,
	}

	// do cancle subscritpion
	err = cardObject.CancleSubscription(requestObj.PaymentIntent)

	// check for an error
	if err != nil {
		log.Println("error when cancling subscription user : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// update status on database
	err = app.db.UpdateOrderStatus(requestObj.ID, 3)

	// check for an error
	if err != nil {
		log.Println("error when updating order status on database")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, errors.New("success to cancle subscription, but failed to update status in database"))
		return
	}

	// if success creating success response
	responseUser := JsonResponse{
		OK:      false, // means does'nt have any error when requesting refund
		Message: "success refunding",
		Content: fmt.Sprintf("success cancle subscription with id : %d", requestObj.ID),
		Id:      uuid.New().String(),
	}

	// send response
	app.WriteJsonObject(w, &responseUser, http.StatusAccepted)
}

// cretae function to get all users
func (app *application) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// get all users from database
	allUsers, err := app.db.GetAllUsers()

	// check for an error
	if err != nil {
		log.Println("error when getting all users from database")
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// if success
	// send all users as response
	app.WriteJsonObject(w, &allUsers, http.StatusAccepted)
}

func (app *application) GetUser(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// get id from routing
	getIdStr := chi.URLParam(r, "id")

	// convert id to integer
	getId, err := strconv.Atoi(getIdStr)

	// check for an error
	if err != nil {
		log.Println("error when converting string id to integer : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// get user from database based id
	getUser, err := app.db.GetUser(getId)

	// check for an error
	if err != nil {
		log.Println("error when getting user from dabase by id : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// send user as response
	app.WriteJsonObject(w, &getUser, http.StatusAccepted)
}

// create function to edit user
func (app *application) EditUser(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// get id from url
	getIdStr := chi.URLParam(r, "id")

	// convert id into integer
	getId, err := strconv.Atoi(getIdStr)

	// check for an error
	if err != nil {
		log.Println("errro when converting id string into ineteger : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create user object
	var userRequest models.User

	// read user data from request
	err = app.ReadJsonBodyRequest(w, r, &userRequest)

	// check for an error
	if err != nil {
		log.Println("errro when read user request : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// check for an id
	if getId != 0 {
		// if user to be edit is old user wnat to be edit
		// edit user
		err = app.db.UpdateUser(userRequest)

		// check for an error
		if err != nil {
			log.Println("errro when updating user in database : ", err)
			app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
			return
		}

		// also check, if user want to update their password or not
		if userRequest.Password != "" {
			// if user wanted to update their password
			// generate hash password from bycrypt
			hashPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), 12)

			// check for an error
			if err != nil {
				log.Println("errro when generate new hash password for updating password : ", err)
				app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
				return
			}

			// update password
			err = app.db.UpdateUserPassword(userRequest, string(hashPassword))

			// check for an error
			if err != nil {
				log.Println("errro when updating user password in database : ", err)
				app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
				return
			}
		}
	} else {
		// if user to be edit is new user and need to be added to database
		// create hash password too
		hashPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), 12)

		// check for an error
		if err != nil {
			log.Println("errro when generate new hash password for updating password : ", err)
			app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
			return
		}

		// add to database
		err = app.db.AddUser(userRequest, string(hashPassword))

		// check for an error
		if err != nil {
			log.Println("errro when adding new user to database : ", err)
			app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	// create response back to user
	responseSuccess := JsonResponse{
		OK:      false, // means no error
		Message: "Success editing user",
		Content: fmt.Sprintf("User editing with id : %d", getId),
		Id:      uuid.New().String(),
	}

	// send response back
	app.WriteJsonObject(w, &responseSuccess, http.StatusAccepted)
}

func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// set header
	w.Header().Set("Content-Type", "application/json")

	// get id from routing
	getIdStr := chi.URLParam(r, "id")

	// convert id to integer
	getId, err := strconv.Atoi(getIdStr)

	// check for an error
	if err != nil {
		log.Println("error when converting string id to integer : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// delete user
	err = app.db.DeleteUser(getId)

	// check for an error
	if err != nil {
		log.Println("error when deleting user in database : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// if success create response
	response := JsonResponse{
		OK:      false,
		Message: "Success deleting user",
		Content: fmt.Sprintf("Deleting user with id : %d", getId),
	}

	// send response
	app.WriteJsonObject(w, &response, http.StatusAccepted)
}
