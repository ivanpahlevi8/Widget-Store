package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

func (app *application) InvoiceEndPoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from invoice endpoint : %s", "Hello")
}

// create order object to reeceiuve payload from request to creating invoice
type OrderPayload struct {
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

func (app *application) CreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	// set requeest as json response
	w.Header().Set("Content-Type", "application/json")

	// create object to hold value from request
	var orderPayload OrderPayload

	// read json object from request
	err := app.ReadJsonBodyRequest(w, r, &orderPayload)

	// check for an error
	if err != nil {
		log.Println("error when creating request to create and send invoice : ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create object input for testing invoice
	// orderPayload.ID = 10
	// orderPayload.Quantity = 5
	// orderPayload.Amount = 5000
	// orderPayload.ProductName = "Widget with RGB Color"
	// orderPayload.FirstName = "Ivan Indirsyah"
	// orderPayload.LastName = "Pahlevi"
	// orderPayload.Email = "ivan.indirsya@gmail.com"
	// orderPayload.Message = "Success creating invoice..."
	// orderPayload.CreatedAt = time.Now()

	// create pdf
	err = app.CreatePdf(orderPayload)

	// check for an error
	if err != nil {
		log.Println("error when creating pdf file based on order object: ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// send email to user
	err = app.SendEmail(
		orderPayload.Email,
		"widget@store.com",
		"Invoice Purchasing",
		"email-attacth",
		[]string{
			fmt.Sprintf("./invoices/%d.pdf", orderPayload.ID),
		},
		nil,
	)

	// check for an error
	if err != nil {
		log.Println("error when sending email to user: ", err)
		app.ErrorJsonResponse(w, http.StatusInternalServerError, err)
		return
	}

	// create response to user
	responsePayload := JsonResponse{
		OK:      false,
		Message: "Success creating and sending invoice",
		Content: fmt.Sprintf("Created invoice %d.pdf and sending to %s", orderPayload.ID, orderPayload.Email),
	}

	// sedn response to user
	app.WriteJsonObject(w, &responsePayload, http.StatusAccepted)
}

// create function to create an pdf
func (a *application) CreatePdf(order OrderPayload) error {
	// create pdf object
	pdf := gofpdf.New("P", "mm", "Letter", "")

	// set pdf margin
	pdf.SetMargins(10, 13, 10)

	// set page breaker on page
	pdf.SetAutoPageBreak(true, 10)

	// create importer object to import pdf templates
	importer := gofpdi.NewImporter()

	// import file pdf formate from ropot directory
	t := importer.ImportPage(pdf, "./pdf-templates/invoice.pdf", 1, "/MediaBox")

	// create one page odf pdf
	pdf.AddPage()

	// user template
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	// create first info
	// set postition to be written
	pdf.SetY(50)
	pdf.SetX(10)
	// set font
	pdf.SetFont("Arial", "", 12)
	// crate cell format to be filled with text with box width 97 and height of box 8 in mm
	pdf.CellFormat(97, 8, fmt.Sprintf("Attention to : Mr.%s, %s", order.LastName, order.FirstName), "", 0, "L", false, 0, "")

	// move cursor to down by 5
	pdf.Ln(5)
	pdf.CellFormat(97, 8, fmt.Sprintf("With Email : %s", order.Email), "", 0, "L", false, 0, "")

	// movbe down by 5 to write date
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.CreatedAt.Format("2006-01-02"), "", 0, "L", false, 0, "")

	// move to new position to write product
	pdf.SetY(93)
	pdf.SetX(58)
	pdf.CellFormat(155, 8, order.ProductName, "", 0, "L", false, 0, "")

	// move to new position to write product quantity
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", order.Quantity), "", 0, "C", false, 0, "")

	// move to new posisiton in same cell to write price
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$ %.2f", float32(order.Amount/100.0)), "", 0, "R", false, 0, "")

	// move down to new position to write another extra product
	pdf.Ln(10)
	pdf.SetX(58)
	pdf.CellFormat(20, 8, "Wrapping Packaging", "", 0, "L", false, 0, "")

	// move to new position to write product quantity
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", 1), "", 0, "C", false, 0, "")

	// move to new posisiton in same cell to write price
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$ %.2f", float32(100/100.0)), "", 0, "R", false, 0, "")

	// move to new position to write new product
	pdf.Ln(10)
	pdf.SetX(58)
	pdf.CellFormat(20, 8, "Order Packaging", "", 0, "L", false, 0, "")

	// move to new position to write product quantity
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", 1), "", 0, "C", false, 0, "")

	// move to new posisiton in same cell to write price
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$ %.2f", float32(200/100.0)), "", 0, "R", false, 0, "")

	// move to new posisiton to write total
	pdf.Ln(125)

	// move to new position to write product quantity
	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", order.Quantity+2), "", 0, "C", false, 0, "")

	// move to new posisiton in same cell to write price
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$ %.2f", float32((order.Amount+300)/100.0)), "", 0, "R", false, 0, "")

	// create route to directory to save new pdf file
	invoiceDir := fmt.Sprintf("./invoices/%d.pdf", order.ID)

	// create pdf file
	err := pdf.OutputFileAndClose(invoiceDir)

	// cheeck for an error
	if err != nil {
		log.Println("error when creating pdf : ", err)
		return err
	}

	return nil
}
