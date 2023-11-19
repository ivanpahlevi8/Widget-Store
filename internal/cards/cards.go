package cards

import (
	"log"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/refund"
	"github.com/stripe/stripe-go/v72/sub"
)

// create struct typoe of card
type Card struct {
	Key      string
	Secret   string
	Currency string
}

// create payment object to hold payments process
type Transaction struct {
	TransactionStatusId int
	Currency            string
	Amount              int
	LastFour            string
	BankCode            string
}

// crete function to charge
/**
function ini digunakan untuk decoupled antaraa pembuatan payment intent dengan process charge
karenea payment intent dapat berasal dari berbagaiu jenis pembayaran dan berbagaiu bank
sehingga harus dilakukan decupled
*/
func (card *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return card.CreatePaymetnIntent(currency, amount)
}

// create function to create payment intent
// payment intent digunbakan untuk memproses pembayaran dari kartu kredit
func (card *Card) CreatePaymetnIntent(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	// assign stripe secrete
	stripe.Key = card.Secret

	// create payment intent value
	paymentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(currency),
	}

	// create payment intent
	paymentIntent, err := paymentintent.New(paymentParams)

	// check for an error
	if err != nil {
		// create msg error from intent
		msg := ""
		// get error from stiripes
		stripeErr, ok := err.(*stripe.Error)
		if ok {
			// if there is an error
			msg = processError(stripeErr)
		}

		// return fail
		return nil, msg, err
	}

	// if success and not have an error
	return paymentIntent, "", nil
}

/**
dua methoid dibawah yang digunakan untuk mendapatkan payment intent dan payment method
dua object ini digunakan untuk melakukan authentikasi terhadap card unutk di charge
karena, tidak mungkin untuk mendapatkan keseluruhan dari credit card number untuk diletakkan di database
payment intent digunakan untuk mendapatkan id dari bank
payment menthod digunakna untuk  mendapatkan last 4 digit creadit card number, and date expired
ketiga data tersebut akan digunakan untk melakukan charge terhadap kartu credit
*/

// create method to get card payment intent
func (card *Card) GetCardPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	// set secrete key
	stripe.Key = card.Secret

	// get intent payment
	pi, err := paymentintent.Get(id, nil)

	// check for an error
	if err != nil {
		log.Println("error when getting payment intent card")
		return nil, err
	}

	// if success
	return pi, nil
}

// cretae method to get payment method
func (card *Card) GetCardPaymentMethod(id string) (*stripe.PaymentMethod, error) {
	// set stripe key
	stripe.Key = card.Secret

	// get payment method
	pm, err := paymentmethod.Get(id, nil)

	// check for an error
	if err != nil {
		log.Println("error when getting payment method")
		return nil, err
	}

	// if success
	return pm, nil
}

// create function to create customer
func (card *Card) CreateCustomer(paymentMethod string, email string) (*stripe.Customer, string, error) {
	//set stripe key to secret key api
	stripe.Key = card.Secret

	// create customer params
	/**
	untuk membuat customer, diperlukan parameter dari customer
	yang dalma hal ini adlaah berkaitan dengan customer seperti email dan metode pembayaran dari customer
	*/
	custParams := &stripe.CustomerParams{
		Email:         stripe.String(email),
		PaymentMethod: stripe.String(paymentMethod),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethod),
		},
	}

	// create customer
	/**
	selanjutnya, dapat membuat object customer berdasarkan parameter yang dipilih pada customer
	sebelumnya
	*/
	newCustomer, err := customer.New(custParams)

	// check for an error
	if err != nil {
		msg := ""
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = processError(stripeErr)
		}
		return nil, msg, err
	}

	// if success
	return newCustomer, "", nil
}

// create function to add subscription
func (card *Card) AddNewSubscription(customer *stripe.Customer, lastFour string, paymentMethod string, email string, cardType string, plan string) (*stripe.Subscription, error) {
	// get customer id
	custId := customer.ID

	// cretae subscription params
	/**
	dalam subscript tentunya terdapat plan atau rencana mengenai pembayaran dan hal yang berkaitan dengan subs
	sehingga, harus membuat parameter dari item susbcriber yang dalma hal ini adalah params
	*/
	subsItemParams := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}

	// create subscription params
	/**
	selanjutnya, setelah subs item params telah dibuat, akan dibuat subs params
	dimana subs params mewakili subscription itu sendiri yang terdiri dari customer yang melakukan
	subscriptiin dan item atau dalam hal ini adalah parameter dari item yang dilakukan subscription
	*/
	subsParams := &stripe.SubscriptionParams{
		Customer: stripe.String(custId),
		Items:    subsItemParams,
	}

	//add meta data to subs params
	/**
	selnajutnya meta data ditambahkan untuk informasi yang akan digunakna pada sisi front end jika
	diperlukan utnuk dipergunakan
	*/
	subsParams.AddMetadata("last_four", lastFour)
	subsParams.AddMetadata("email", email)
	subsParams.AddMetadata("card_type", cardType)
	subsParams.AddExpand("latest_invoice.payment_intent")

	// create subscription
	/**
	selanjutnya, dapat membuat subscription object berdasarkan subscription parameter
	*/
	subs, err := sub.New(subsParams)

	// check for an error
	if err != nil {
		log.Println("error when createing subscribe : ", err)
		return nil, err
	}

	// if success
	return subs, nil
}

// create function to refund payment of cards
func (c *Card) RefundPayment(pi string, amount int) error {
	// set secrete key
	stripe.Key = c.Secret

	// convert amount of integer to int64
	amountGet := int64(amount)

	// create refunds params
	refundParams := &stripe.RefundParams{
		PaymentIntent: &pi,
		Amount:        &amountGet,
	}

	// create refund object based on refund params
	_, err := refund.New(refundParams)

	// check for an error
	if err != nil {
		log.Println("error whenb creating refund object : ", err)
		return err
	}

	// if success
	return nil
}

// create function to cancle subscription
func (c *Card) CancleSubscription(subsId string) error {
	// set secret key
	stripe.Key = c.Secret

	// create subscription params
	subsParams := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true), // cancle user at its end period
	}

	// update subscription
	_, err := sub.Update(subsId, subsParams)

	// check for an error
	if err != nil {
		log.Println("error when cancling user subscription")
		return err
	}

	// if success
	return nil
}

// create function to processing error
func processError(a *stripe.Error) string {
	// get error code
	errorCode := a.Code

	// create string result
	res := ""

	switch errorCode {
	case stripe.ErrorCodeCardDeclined:
		res = "stripe error card declined"
	case stripe.ErrorCodeBankAccountDeclined:
		res = "stripe error account declined"
	case stripe.ErrorCodeExpiredCard:
		res = "stripe error card expired"
	default:
		res = "error when processing card"
	}

	return res
}
