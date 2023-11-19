package models

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// create model that hold database connection
type DbModel struct {
	DbConn *sql.DB
}

// create db model that hold model
type Model struct {
	Db DbModel
}

// create function to create DbModel object
func NewDbModel(dbConn *sql.DB) Model {
	// create new model
	myModel := DbModel{
		DbConn: dbConn,
	}

	// crete DbModel
	dbModel := Model{
		Db: myModel,
	}

	// return value
	return dbModel
}

// create type that hold widget value
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsReccuring    bool      `json:"is_reccuring"`
	PlanId         string    `json:"plan_id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

// create model to hold orders object
type Order struct {
	ID               int         `json:"id"`
	WidgetID         int         `json:"widget_id"`
	TransactionID    int         `json:"transaction_id"`
	StatusID         int         `json:"status_id"`
	CustomersID      int         `json:"customer_id"`
	Quantity         int         `json:"quantity"`
	Amount           int         `json:"amount"`
	CreatedAt        time.Time   `json:"-"`
	UpdatedAt        time.Time   `json:"-"`
	TransactionModel Transaction `json:"transaction_model"`
	WidgetModel      Widget      `json:"widget_model"`
	CustomerModel    Customer    `json:"customer_model"`
}

// create model to hold statuses
type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// create model to hold transaction
type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusId int       `json:"transaction_status_id"`
	ExpiredMonth        int       `json:"expired_month"`
	ExpiredYear         int       `json:"expired_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

// create model to hold transaction status
type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// create function to hold user
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// create model to hold customers
type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// create funciton to access database
func (db *DbModel) GetWidgetById(id int) (Widget, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// cretate variable object
	var getId int
	var getName string
	var getDescription string
	var getInventoryLevel int
	var getPrice int
	var getCreatedAt time.Time
	var getUpdatedAt time.Time
	var getImage string
	var getIsReccuring bool
	var getPlanId string

	// create object
	var objReturn Widget

	// create query syntax
	querySyntax := `
	SELECT id, name, description, inventory_level, price, created_at, updated_at, image, is_reccuring, plan_id
	FROM widgets WHERE id=?`

	// create query
	query := db.DbConn.QueryRowContext(ctx, querySyntax, id)

	// scan query
	err := query.Scan(
		&getId,
		&getName,
		&getDescription,
		&getInventoryLevel,
		&getPrice,
		&getCreatedAt,
		&getUpdatedAt,
		&getImage,
		&getIsReccuring,
		&getPlanId,
	)

	// check for an error
	if err != nil {
		log.Println("error when querying to get widget by id")
		return objReturn, err
	}

	// assign value to object
	objReturn.ID = getId
	objReturn.Name = getName
	objReturn.Description = getDescription
	objReturn.InventoryLevel = getInventoryLevel
	objReturn.Price = getPrice
	objReturn.CreatedAt = getCreatedAt
	objReturn.UpdatedAt = getUpdatedAt
	objReturn.IsReccuring = getIsReccuring
	objReturn.PlanId = getPlanId

	// return object
	return objReturn, nil
}

// create function to insert transaction
func (db *DbModel) InsertTransaction(trx Transaction) (int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create query text
	queryTxt := `
		INSERT INTO transactions(
			amount, currency, last_four, bank_return_code, transaction_status_id, 
			created_at, updated_at, expired_month, expired_year, payment_intent, payment_method
		)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// execute query
	res, err := db.DbConn.ExecContext(
		ctx,
		queryTxt,
		trx.Amount,
		trx.Currency,
		trx.LastFour,
		trx.BankReturnCode,
		trx.TransactionStatusId,
		trx.CreatedAt,
		trx.UpdatedAt,
		trx.ExpiredMonth,
		trx.ExpiredYear,
		trx.PaymentIntent,
		trx.PaymentMethod,
	)

	// check for an error
	if err != nil {
		log.Println("error when executing query to add transaction to database")
		return -1, err
	}

	// get id as feedback
	id, err := res.LastInsertId()

	// check for an error
	if err != nil {
		log.Println("error when getting id transaction to database")
		return -1, err
	}

	// if success
	return int(id), nil
}

// create function to insert order
func (db *DbModel) InsertOrder(order Order) (int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create query text
	queryTxt := `
		INSERT INTO orders(widget_id, transaction_id, status_id, quantity, amount, created_at, updated_at, customers_id)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
	`

	// execute query
	res, err := db.DbConn.ExecContext(
		ctx,
		queryTxt,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		order.CreatedAt,
		order.UpdatedAt,
		order.CustomersID,
	)

	// check for an error
	if err != nil {
		log.Println("error when executing query to add transaction to database")
		return -1, err
	}

	// get id as feedback
	id, err := res.LastInsertId()

	// check for an error
	if err != nil {
		log.Println("error when getting id transaction to database")
		return -1, err
	}

	// if success
	return int(id), nil
}

// create function to insert customers
func (db *DbModel) InsertCustomer(c Customer) (int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create query text
	queryTxt := `
		INSERT INTO customers(first_name, last_name, email, updated_at, created_at)
		VALUES(?, ?, ?, ?, ?)
	`

	// execute query
	res, err := db.DbConn.ExecContext(
		ctx,
		queryTxt,
		c.FirstName,
		c.LastName,
		c.Email,
		c.UpdatedAt,
		c.CreatedAt,
	)

	// check for an error
	if err != nil {
		log.Println("error when executing query to add transaction to database")
		return -1, err
	}

	// get id as feedback
	id, err := res.LastInsertId()

	// check for an error
	if err != nil {
		log.Println("error when getting id transaction to database")
		return -1, err
	}

	// if success
	return int(id), nil
}

// create function to get user by email
func (db *DbModel) GetUserByEmail(email string) (User, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create variable to hold data from database
	var getId int
	var getFirstName string
	var getLastName string
	var getEmail string
	var getPassword string
	var getCreatedAt time.Time
	var getUpdatedAt time.Time

	// create object to hold user
	var user User

	// create query
	queryText := `
		SELECT id, first_name, last_name, email, password, created_at, updated_at FROM users
		WHERE email=?
	`

	// query database
	query := db.DbConn.QueryRowContext(ctx, queryText, email)

	// scan query to assign variable
	err := query.Scan(
		&getId,
		&getFirstName,
		&getLastName,
		&getEmail,
		&getPassword,
		&getCreatedAt,
		&getUpdatedAt,
	)

	// check for an error
	if err != nil {
		log.Println("error when query to get user by email : ", err)
		return user, err
	}

	// create object user
	user = User{
		ID:        getId,
		FirstName: getFirstName,
		LastName:  getLastName,
		Email:     getEmail,
		Password:  getPassword,
		CreatedAt: getCreatedAt,
		UpdatedAt: getUpdatedAt,
	}

	// check for query error
	err = query.Err()

	if err != nil {
		log.Println("error when query to get user by email 2: ", err)
		return user, err
	}

	// if success
	return user, nil
}

// create function to validate user in front end
func (db *DbModel) ValidateUser(email string, password string) (int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create variable to hold value from database
	var getId int
	var getHashPassword string

	// create query
	queryTxt := `
		SELECT id, password FROM users WHERE email=?
	`

	// do query
	query := db.DbConn.QueryRowContext(
		ctx,
		queryTxt,
		email,
	)

	// scan query
	err := query.Scan(
		&getId,
		&getHashPassword,
	)

	// check for an error
	if err != nil {
		log.Println("error when trying to getting user from database using email")
		return -1, err
	}

	// compare password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(getHashPassword), []byte(password))

	// check for an error
	if err == bcrypt.ErrMismatchedHashAndPassword {
		// if error because usier inputing invalid password
		log.Println("user input non valid password credentials")
		return -1, errors.New("invalid user password")
	} else if err != nil {
		log.Println("error happend : ", err)
		return -1, err
	}

	// if authentication success
	return getId, nil
}

// create function to upodate user password with new password
func (db *DbModel) UpdateUserPassword(user User, passwordHash string) error {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// get variable id from user
	getId := user.ID

	// crete query
	queryStmt := `
		UPDATE users SET password=? WHERE id=?
	`

	// start query
	_, err := db.DbConn.ExecContext(
		ctx,
		queryStmt,
		passwordHash,
		getId,
	)

	// check for an error
	if err != nil {
		log.Println(err)
		return err
	}

	// if success
	return nil
}

// create function to get all orders
func (db *DbModel) GetAllOrders() ([]Order, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create order, transaction, widget, and customer model to hold value of model
	var orderObject Order
	var transactionObject Transaction
	var widgetObject Widget
	var customerObject Customer

	// create array to hold reference to order
	var allOrders []Order

	// cretae quer
	queryTxt := `
	SELECT
		o.id, o.widget_id, o.transaction_id,o.status_id, o.quantity, o.amount, o.created_at, o.updated_at, o.customers_id,
		w.id, w.name, w.description, w.price,
		t.amount, t.currency, t.last_four, t.bank_return_code, t.expired_month, t.expired_year, t.payment_intent, t.payment_method,
		c.first_name, c.last_name, c.email
	FROM
		orders o
		INNER JOIN widgets w ON (o.widget_id = w.id)
		INNER JOIN transactions t ON (o.transaction_id = t.id)
		INNER JOIN customers c ON (o.customers_id = c.id)
	WHERE
		w.is_reccuring = 0
	ORDER BY
		o.created_at desc	
	`

	// query text
	query, err := db.DbConn.QueryContext(
		ctx,
		queryTxt,
	)

	// check for an error
	if err != nil {
		log.Println("error when querying to get all orders")
		return nil, err
	}

	// scan query
	for query.Next() {
		err = query.Scan(
			&orderObject.ID,
			&orderObject.WidgetID,
			&orderObject.TransactionID,
			&orderObject.StatusID,
			&orderObject.Quantity,
			&orderObject.Amount,
			&orderObject.CreatedAt,
			&orderObject.UpdatedAt,
			&orderObject.CustomersID,
			&widgetObject.ID,
			&widgetObject.Name,
			&widgetObject.Description,
			&widgetObject.Price,
			&transactionObject.Amount,
			&transactionObject.Currency,
			&transactionObject.LastFour,
			&transactionObject.BankReturnCode,
			&transactionObject.ExpiredMonth,
			&transactionObject.ExpiredYear,
			&transactionObject.PaymentIntent,
			&transactionObject.PaymentMethod,
			&customerObject.FirstName,
			&customerObject.LastName,
			&customerObject.Email,
		)

		// check for an error
		if err != nil {
			log.Println("error when scanning data from database to object : ", err)
			return nil, err
		}

		// assign transaction, widget, and customer to order
		orderObject.TransactionModel = transactionObject
		orderObject.WidgetModel = widgetObject
		orderObject.CustomerModel = customerObject

		// assignt order object to slicae
		allOrders = append(allOrders, orderObject)
	}

	// check for an eror
	err = query.Err()
	if err != nil {
		log.Println("error when scanning data : ", err)
		return nil, err
	}

	// if success
	return allOrders, nil
}

// create function to get all orders paginated
func (db *DbModel) GetAllOrdersPaginated(limitPage int, page int) ([]Order, int, int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// get offset page
	offsetPage := (page - 1) * limitPage
	/**
	offset page merupakan component terakhir yang terdapat pada page, misalnua pada page terdapat 10 object
	maka offset dari page tersebut adalah sebesar 10. Sehingga informasi ini dapat digunakan untuk
	menampilkan mulai dari data keberapa pada halaman berikutnya
	*/

	// create order, transaction, widget, and customer model to hold value of model
	var orderObject Order
	var transactionObject Transaction
	var widgetObject Widget
	var customerObject Customer

	// create array to hold reference to order
	var allOrders []Order

	// cretae quer
	queryTxt := `
	SELECT
		o.id, o.widget_id, o.transaction_id,o.status_id, o.quantity, o.amount, o.created_at, o.updated_at, o.customers_id,
		w.id, w.name, w.description, w.price,
		t.amount, t.currency, t.last_four, t.bank_return_code, t.expired_month, t.expired_year, t.payment_intent, t.payment_method,
		c.first_name, c.last_name, c.email
	FROM
		orders o
		INNER JOIN widgets w ON (o.widget_id = w.id)
		INNER JOIN transactions t ON (o.transaction_id = t.id)
		INNER JOIN customers c ON (o.customers_id = c.id)
	WHERE
		w.is_reccuring = 0
	ORDER BY
		o.created_at desc
	limit ? offset ?
	`

	// query text
	query, err := db.DbConn.QueryContext(
		ctx,
		queryTxt,
		limitPage,  // akan memberikan data sebanyak limit yang diberikan pada query
		offsetPage, // data yang diberikan dimulai dari offset keberapa
	)
	/**
	misalkan limitPage = 5 dan offsetPage = 3, maka query akan memberikan data sebanyak 5 data dan dimulai
	dari data ke tiga.
	*/

	// check for an error
	if err != nil {
		log.Println("error when querying to get all orders")
		return nil, 0, 0, err
	}

	// scan query
	for query.Next() {
		err = query.Scan(
			&orderObject.ID,
			&orderObject.WidgetID,
			&orderObject.TransactionID,
			&orderObject.StatusID,
			&orderObject.Quantity,
			&orderObject.Amount,
			&orderObject.CreatedAt,
			&orderObject.UpdatedAt,
			&orderObject.CustomersID,
			&widgetObject.ID,
			&widgetObject.Name,
			&widgetObject.Description,
			&widgetObject.Price,
			&transactionObject.Amount,
			&transactionObject.Currency,
			&transactionObject.LastFour,
			&transactionObject.BankReturnCode,
			&transactionObject.ExpiredMonth,
			&transactionObject.ExpiredYear,
			&transactionObject.PaymentIntent,
			&transactionObject.PaymentMethod,
			&customerObject.FirstName,
			&customerObject.LastName,
			&customerObject.Email,
		)

		// check for an error
		if err != nil {
			log.Println("error when scanning data from database to object : ", err)
			return nil, 0, 0, err
		}

		// assign transaction, widget, and customer to order
		orderObject.TransactionModel = transactionObject
		orderObject.WidgetModel = widgetObject
		orderObject.CustomerModel = customerObject

		// assignt order object to slicae
		allOrders = append(allOrders, orderObject)
	}

	// check for an eror
	err = query.Err()
	if err != nil {
		log.Println("error when scanning data : ", err)
		return nil, 0, 0, err
	}

	// get data left
	var dataOffsetLeft int

	queryTxt = `
		SELECT count(o.id)
		FROM orders o
		LEFT JOIN widgets w on (o.widget_id = w.id)
		WHERE w.is_reccuring = 0
	`

	// execute query
	rowSelect := db.DbConn.QueryRowContext(ctx, queryTxt)

	// get value
	err = rowSelect.Scan(&dataOffsetLeft)

	// check for an error
	if err != nil {
		log.Println("error when getting all data set in database")
	}

	// get last page
	lastPage := dataOffsetLeft / limitPage
	/**
	akan mereturn nilai dari banyak halaman, dimana nilai banyak halaman merepresetnasikan
	nilai dari halaman terakhir
	*/

	// if success
	return allOrders, dataOffsetLeft, lastPage, nil
}

// create function to get all subscription
func (db *DbModel) GetAllSubscription() ([]Order, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create order, transaction, widget, and customer model to hold value of model
	var orderObject Order
	var transactionObject Transaction
	var widgetObject Widget
	var customerObject Customer

	// create array to hold reference to order
	var allOrders []Order

	// cretae quer
	queryTxt := `
	SELECT
		o.id, o.widget_id, o.transaction_id,o.status_id, o.quantity, o.amount, o.created_at, o.updated_at, o.customers_id,
		w.id, w.name, w.description, w.price,
		t.amount, t.currency, t.last_four, t.bank_return_code, t.expired_month, t.expired_year, t.payment_intent, t.payment_method,
		c.first_name, c.last_name, c.email
	FROM
		orders o
		INNER JOIN widgets w ON (o.widget_id = w.id)
		INNER JOIN transactions t ON (o.transaction_id = t.id)
		INNER JOIN customers c ON (o.customers_id = c.id)
	WHERE
		w.is_reccuring = 1
	ORDER BY
		o.created_at desc	
	`

	// query text
	query, err := db.DbConn.QueryContext(
		ctx,
		queryTxt,
	)

	// check for an error
	if err != nil {
		log.Println("error when querying to get all orders")
		return nil, err
	}

	// scan query
	for query.Next() {
		err = query.Scan(
			&orderObject.ID,
			&orderObject.WidgetID,
			&orderObject.TransactionID,
			&orderObject.StatusID,
			&orderObject.Quantity,
			&orderObject.Amount,
			&orderObject.CreatedAt,
			&orderObject.UpdatedAt,
			&orderObject.CustomersID,
			&widgetObject.ID,
			&widgetObject.Name,
			&widgetObject.Description,
			&widgetObject.Price,
			&transactionObject.Amount,
			&transactionObject.Currency,
			&transactionObject.LastFour,
			&transactionObject.BankReturnCode,
			&transactionObject.ExpiredMonth,
			&transactionObject.ExpiredYear,
			&transactionObject.PaymentIntent,
			&transactionObject.PaymentMethod,
			&customerObject.FirstName,
			&customerObject.LastName,
			&customerObject.Email,
		)

		// check for an error
		if err != nil {
			log.Println("error when scanning data from database to object : ", err)
			return nil, err
		}

		// assign transaction, widget, and customer to order
		orderObject.TransactionModel = transactionObject
		orderObject.WidgetModel = widgetObject
		orderObject.CustomerModel = customerObject

		// assignt order object to slicae
		allOrders = append(allOrders, orderObject)
	}

	// check for an eror
	err = query.Err()
	if err != nil {
		log.Println("error when scanning data : ", err)
		return nil, err
	}

	// if success
	return allOrders, nil
}

// create function to get all subscription pagoinated
func (db *DbModel) GetAllSubscriptionPaginated(pageLimit int, currPage int) ([]Order, int, int, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create variable to show last component in current page
	currComponent := (currPage - 1) * pageLimit

	// create order, transaction, widget, and customer model to hold value of model
	var orderObject Order
	var transactionObject Transaction
	var widgetObject Widget
	var customerObject Customer

	// create array to hold reference to order
	var allOrders []Order

	// cretae quer
	queryTxt := `
	SELECT
		o.id, o.widget_id, o.transaction_id,o.status_id, o.quantity, o.amount, o.created_at, o.updated_at, o.customers_id,
		w.id, w.name, w.description, w.price,
		t.amount, t.currency, t.last_four, t.bank_return_code, t.expired_month, t.expired_year, t.payment_intent, t.payment_method,
		c.first_name, c.last_name, c.email
	FROM
		orders o
		INNER JOIN widgets w ON (o.widget_id = w.id)
		INNER JOIN transactions t ON (o.transaction_id = t.id)
		INNER JOIN customers c ON (o.customers_id = c.id)
	WHERE
		w.is_reccuring = 1
	ORDER BY
		o.created_at desc
	limit ? offset ?	
	`

	// query text
	query, err := db.DbConn.QueryContext(
		ctx,
		queryTxt,
		pageLimit,
		currComponent,
	)

	// check for an error
	if err != nil {
		log.Println("error when querying to get all orders")
		return nil, 0, 0, err
	}

	// scan query
	for query.Next() {
		err = query.Scan(
			&orderObject.ID,
			&orderObject.WidgetID,
			&orderObject.TransactionID,
			&orderObject.StatusID,
			&orderObject.Quantity,
			&orderObject.Amount,
			&orderObject.CreatedAt,
			&orderObject.UpdatedAt,
			&orderObject.CustomersID,
			&widgetObject.ID,
			&widgetObject.Name,
			&widgetObject.Description,
			&widgetObject.Price,
			&transactionObject.Amount,
			&transactionObject.Currency,
			&transactionObject.LastFour,
			&transactionObject.BankReturnCode,
			&transactionObject.ExpiredMonth,
			&transactionObject.ExpiredYear,
			&transactionObject.PaymentIntent,
			&transactionObject.PaymentMethod,
			&customerObject.FirstName,
			&customerObject.LastName,
			&customerObject.Email,
		)

		// check for an error
		if err != nil {
			log.Println("error when scanning data from database to object : ", err)
			return nil, 0, 0, err
		}

		// assign transaction, widget, and customer to order
		orderObject.TransactionModel = transactionObject
		orderObject.WidgetModel = widgetObject
		orderObject.CustomerModel = customerObject

		// assignt order object to slicae
		allOrders = append(allOrders, orderObject)
	}

	// check for an eror
	err = query.Err()
	if err != nil {
		log.Println("error when scanning data : ", err)
		return nil, 0, 0, err
	}

	// get all subscription in database
	var allSubscription int

	// statement
	stmt := `
		SELECT count(o.id)
		FROM orders o
		LEFT JOIN widgets w ON(o.widget_id = w.id)
		WHERE w.is_reccuring = 1
	`

	// scan statement
	queryStmt := db.DbConn.QueryRowContext(ctx, stmt)

	// assign value
	err = queryStmt.Scan(&allSubscription)

	// check for an error
	if err != nil {
		log.Println("error when counting all data of subscription : ", err)
		return nil, 0, 0, err
	}

	allPages := allSubscription / pageLimit

	// if success
	return allOrders, allSubscription, allPages, nil
}

// cretae function to get order by id
func (db *DbModel) GetOrderById(id int) (Order, error) {
	// create context
	// with timeout 10 second
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// defer close
	defer cancle()

	// create order, transaction, widget, and customer model to hold value of model
	var orderObject Order
	var transactionObject Transaction
	var widgetObject Widget
	var customerObject Customer

	// cretae quer
	queryTxt := `
	SELECT
		o.id, o.widget_id, o.transaction_id,o.status_id, o.quantity, o.amount, o.created_at, o.updated_at, o.customers_id,
		w.id, w.name, w.description, w.price,
		t.amount, t.currency, t.last_four, t.bank_return_code, t.expired_month, t.expired_year, t.payment_intent, t.payment_method,
		c.first_name, c.last_name, c.email
	FROM
		orders o
		INNER JOIN widgets w ON (o.widget_id = w.id)
		INNER JOIN transactions t ON (o.transaction_id = t.id)
		INNER JOIN customers c ON (o.customers_id = c.id)
	WHERE
		o.id = ?
	`

	// query text
	query := db.DbConn.QueryRowContext(
		ctx,
		queryTxt,
		id,
	)

	// scan query
	err := query.Scan(
		&orderObject.ID,
		&orderObject.WidgetID,
		&orderObject.TransactionID,
		&orderObject.StatusID,
		&orderObject.Quantity,
		&orderObject.Amount,
		&orderObject.CreatedAt,
		&orderObject.UpdatedAt,
		&orderObject.CustomersID,
		&widgetObject.ID,
		&widgetObject.Name,
		&widgetObject.Description,
		&widgetObject.Price,
		&transactionObject.Amount,
		&transactionObject.Currency,
		&transactionObject.LastFour,
		&transactionObject.BankReturnCode,
		&transactionObject.ExpiredMonth,
		&transactionObject.ExpiredYear,
		&transactionObject.PaymentIntent,
		&transactionObject.PaymentMethod,
		&customerObject.FirstName,
		&customerObject.LastName,
		&customerObject.Email,
	)

	// check for an error
	if err != nil {
		log.Println("error when scanning data from database to object : ", err)
		return Order{}, err
	}

	// assign transaction, widget, and customer to order
	orderObject.TransactionModel = transactionObject
	orderObject.WidgetModel = widgetObject
	orderObject.CustomerModel = customerObject

	// check for an eror
	err = query.Err()
	if err != nil {
		log.Println("error when scanning data : ", err)
		return Order{}, err
	}

	// if success
	return orderObject, nil
}

// create function to update orders status id
func (db *DbModel) UpdateOrderStatus(id int, status int) error {
	// create context
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// close context
	defer cancle()

	// create statement
	stmt := `UPDATE orders SET status_id=? WHERE id=?`

	// execute query
	_, err := db.DbConn.ExecContext(ctx, stmt, status, id)

	// check for an error
	if err != nil {
		log.Println("error when updating status order in database : ", err)
		return err
	}

	// if success
	return nil
}

// create function to get all users from database
func (db *DbModel) GetAllUsers() ([]*User, error) {
	// create contest
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// cancle context
	defer cancle()

	// create variable to hold data
	var allUsers []*User

	// create variable to hold user
	var getId int
	var getFirstName, getLastName, getEmail, getPassword string
	var getCreatedAt, getUpdated_at time.Time

	// create statement
	stmtQuery := `
		SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		ORDER BY first_name, last_name
	`

	// execute query
	query, err := db.DbConn.QueryContext(ctx, stmtQuery)

	// check for an error
	if err != nil {
		log.Println("error when query to get all users")
		return nil, err
	}

	// looping through all query
	for query.Next() {
		err = query.Scan(
			&getId,
			&getFirstName,
			&getLastName,
			&getEmail,
			&getPassword,
			&getCreatedAt,
			&getUpdated_at,
		)

		// check for an error
		if err != nil {
			log.Println("error in looping through all user data")
			return nil, err
		}

		// create object user
		userObj := User{
			ID:        getId,
			FirstName: getFirstName,
			LastName:  getLastName,
			Email:     getEmail,
			Password:  getPassword,
			CreatedAt: getCreatedAt,
			UpdatedAt: getUpdated_at,
		}

		// add user obj to data list
		allUsers = append(allUsers, &userObj)
	}

	// check error from query
	err = query.Err()
	if err != nil {
		log.Println("error when checking querying error")
		return nil, err
	}

	// if success
	return allUsers, nil
}

// create function to get single user
func (db *DbModel) GetUser(id int) (*User, error) {
	// create contest
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// cancle context
	defer cancle()

	// create variable to hold user
	var getId int
	var getFirstName, getLastName, getEmail, getPassword string
	var getCreatedAt, getUpdated_at time.Time

	// create statement
	stmtQuery := `
		SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE id=?
	`

	// execute query
	query := db.DbConn.QueryRowContext(ctx, stmtQuery, id)

	err := query.Scan(
		&getId,
		&getFirstName,
		&getLastName,
		&getEmail,
		&getPassword,
		&getCreatedAt,
		&getUpdated_at,
	)

	// check for an error
	if err != nil {
		log.Println("error in looping through all user data")
		return nil, err
	}

	// create object user
	userObj := User{
		ID:        getId,
		FirstName: getFirstName,
		LastName:  getLastName,
		Email:     getEmail,
		Password:  getPassword,
		CreatedAt: getCreatedAt,
		UpdatedAt: getUpdated_at,
	}

	// check error from query
	err = query.Err()
	if err != nil {
		log.Println("error when checking querying error")
		return nil, err
	}

	// if success
	return &userObj, nil
}

// create function to update users
func (db *DbModel) UpdateUser(newUser User) error {
	// create contest
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// cancle context
	defer cancle()

	// create statement
	stmt := `
		UPDATE users
		SET first_name=?, last_name=?, email=?, updated_at=?
		WHERE id=?
	`

	// execute row
	_, err := db.DbConn.ExecContext(
		ctx,
		stmt,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		time.Now(),
		newUser.ID,
	)

	// check for an error
	if err != nil {
		log.Println("error when updating user by id in database")
		return err
	}

	// if success
	return nil
}

// cretae fucntion to add user with hash password
func (db *DbModel) AddUser(newUser User, hashPass string) error {
	// create contest
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// cancle context
	defer cancle()

	// create statement
	stmt := `
		INSERT INTO users (first_name, last_name, email, password, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// execute query statement
	_, err := db.DbConn.ExecContext(
		ctx,
		stmt,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		hashPass,
		time.Now(),
		time.Now(),
	)

	// check for an error
	if err != nil {
		log.Println("error when adding user by id in database")
		return err
	}

	// if success
	return nil
}

// create function to delete user on database
func (db *DbModel) DeleteUser(id int) error {
	// create contest
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)

	// cancle context
	defer cancle()

	// create statement
	stmt := `
		DELETE FROM users
		WHERE id=?
	`

	// execute statment
	_, err := db.DbConn.ExecContext(ctx, stmt, id)

	// check for an error
	if err != nil {
		log.Println("error when deleting user by id in database")
		return err
	}

	// create function to delete user tokens in database

	// create statement
	stmt = `
		DELETE FROM tokens
		WHERE user_id=?
	`

	// execute statment
	_, err = db.DbConn.ExecContext(ctx, stmt, id)

	// check for an error
	if err != nil {
		log.Println("error when deleting user tokens by id in database")
		return err
	}

	// if success
	return nil
}
