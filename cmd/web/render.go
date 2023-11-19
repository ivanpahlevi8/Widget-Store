package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

/**
buat suatu struct untuk menyimpan data yang akan dipassing ke html template
sehingga, html template akan bersifat dynamic dengan data yang dapat dipassing ke template
*/

type TemplateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	BoolMap         map[string]bool
	DataMap         map[string]interface{}
	CSRFToken       string
	Flash           string
	Error           string
	IsAuthenticated int
	UserID          int
	API             string
}

// create function for template
/**
function ini digunakna agar template dapat mengakses logic bussiness yang diassign
untuk keperluan modifikasi data pada template
*/
var tempFunc = template.FuncMap{
	"add": func(a int, b int) int {
		res := a + b
		return res
	},
	"fmt": formatPrice,
}

// create function to be accessed in template
func formatPrice(n int) string {
	// gte float
	getFloat := float32(n / 100)
	// return string
	return fmt.Sprintf("$%.2f", getFloat)
}

//go:embed template
var tempFS embed.FS

// create function to generate default data to be passed to template html
func (app *Application) DefaultData(td *TemplateData, r *http.Request) *TemplateData {
	// check if there is user id in session or not
	isExist := app.Session.Exists(r.Context(), "user_id")

	if isExist {
		// if user already login
		log.Println("user already login")
		td.IsAuthenticated = 1
		td.UserID = app.Session.GetInt(r.Context(), "user_id")
	} else {
		log.Println("user not yet login")
		td.IsAuthenticated = 0
		td.UserID = 0
	}

	return td
}

// create function to render template
func (app *Application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *TemplateData, partials ...string) error {
	// create object to hold template
	var t *template.Template

	// create object to hold an error
	var err error

	// create variabvle as a source of url
	var templateToRender = fmt.Sprintf("template/%s.page.tmpl", page)

	// check if template data already assigned or not
	if td == nil {
		// if nill assign tremplate data
		td = &TemplateData{}
	}

	// add default data to template data
	td = app.DefaultData(td, r)

	// get template from cache
	templateCache := app.tc[templateToRender]

	// check if it;s in development or production
	if app.config.env == "production" && templateCache != nil {
		// in production
		// and using template cache
		// get template from cache based on page
		t = templateCache
	} else {
		// call function to render template
		t, err = app.parseTemplate(partials, page, templateToRender)
	}

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when parsing template in html template : %s\n", err.Error())
		return err
	}

	// if success
	err = t.Execute(w, &td)

	// check for an error again
	if err != nil {
		app.errorLog.Printf("error when executing the template : %s\n", err.Error())
		return err
	}

	return nil
}

// create function to render template
func (app *Application) parseTemplate(partials []string, page string, templateToRender string) (*template.Template, error) {
	// create object to hold template
	var t *template.Template

	// create object to hold an error
	var err error

	// adding route url to partials if partials exist
	if len(partials) > 0 {
		for i, x := range partials {
			partials[i] = fmt.Sprintf("template/%s.partial.tmpl", x)
		}
	}

	// if partials exist
	if len(partials) > 0 {
		// create template
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(tempFunc).ParseFS(tempFS, "template/base.layout.tmpl", strings.Join(partials, ","), templateToRender)
		// name parameter in new is arbitrary, doenst affect in parsing process
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.tmpl", page)).Funcs(tempFunc).ParseFS(tempFS, "template/base.layout.tmpl", templateToRender)
	}

	// check for an error
	if err != nil {
		app.errorLog.Printf("error when parsing template in html template : %s\n", err.Error())
		return nil, err
	}

	// if nothing error
	return t, nil
}
