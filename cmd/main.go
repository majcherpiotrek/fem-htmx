package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Contact struct {
	Name  string
	Email string
}

type Contacts = []Contact

type FormState struct {
	Values map[string]string
	Errors map[string]string
}

func newFormState() FormState {
	return FormState{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

func (formState *FormState) setFieldError(field, errorMessage string) *FormState {
	formState.Errors[field] = errorMessage

	return formState
}

func (formState *FormState) setFieldValue(field, value string) *FormState {
	formState.Values[field] = value

	return formState
}

func (formState *FormState) hasErrors() bool {
	return len(formState.Errors) > 0
}

type AppState struct {
	Contacts  Contacts
	FormState FormState
}

func (appState *AppState) hasContact(c *Contact) bool {
	for _, contact := range appState.Contacts {
		if contact.Email == c.Email {
			return true
		}

	}
	return false
}

func newContact(name string, email string) Contact {
	return Contact{name, email}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.INFO)

	appState := AppState{
		Contacts:  Contacts{},
		FormState: newFormState(),
	}
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", appState)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		contact := newContact(name, email)

		appState.FormState = newFormState()
		appState.FormState.setFieldValue("name", name)
		appState.FormState.setFieldValue("email", email)

		if appState.hasContact(&contact) {
			appState.FormState.setFieldError("email", "A user with this email already exists")
		}

		if len(name) > 10 {
			appState.FormState.setFieldError("name", "Max length of a name is 10")
		}

		if appState.FormState.hasErrors() {
			return c.Render(422, "contactForm", appState)
		}

		appState.FormState = newFormState()

		appState.Contacts = append(appState.Contacts, contact)

		err := c.Render(200, "oob-contacts", appState)

		if err != nil {
			return c.HTML(500, "Something went wrong")
		}
		return c.Render(200, "contactForm", appState)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
