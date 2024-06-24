package main

import (
	"html/template"
	"io"
	"strconv"

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

var id = 0

type Contact struct {
	Id    int
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

func (appState *AppState) indexOfContactWithId(id int) int {
	for i, contact := range appState.Contacts {
		if contact.Id == id {
			return i
		}

	}
	return -1
}

func newContact(id int, name, email string) Contact {
	return Contact{id, name, email}
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

		contact := newContact(id, name, email)

		appState.FormState = newFormState()
		appState.FormState.setFieldValue("name", name)
		appState.FormState.setFieldValue("email", email)

		if appState.hasContact(&contact) {
			appState.FormState.setFieldError("email", "A user with this email already exists")
		}

		if len(name) > 80 {
			appState.FormState.setFieldError("name", "Max length of a name is 10")
		}

		if appState.FormState.hasErrors() {
			return c.Render(422, "contactForm", appState)
		}

		appState.FormState = newFormState()

		appState.Contacts = append(appState.Contacts, contact)

		err := c.Render(200, "new-contact", contact)

		if err != nil {
			return c.HTML(500, "Something went wrong")
		}

		id++

		err = c.Render(200, "contactsCount", appState.Contacts)

		if err != nil {
			return c.HTML(500, "Something went wrong")
		}

		return c.Render(200, "contactForm", appState)
	})

	e.GET("/contacts", func(c echo.Context) error {
		err := c.Render(200, "contactsCount", appState.Contacts)

		if err != nil {
			return c.HTML(500, "Something went wrong")
		}
		return c.Render(200, "contacts", appState)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)

		if err != nil {
			return c.HTML(400, "Invalid id parameter")
		}

		indexOfContact := appState.indexOfContactWithId(id)

		if indexOfContact == -1 {
			return c.HTML(404, "Contact with this id not found")
		}

		appState.Contacts = append(appState.Contacts[:indexOfContact], appState.Contacts[indexOfContact+1:]...)

		err = c.Render(200, "contactsCount", appState.Contacts)

		if err != nil {
			return c.HTML(500, "Something went wrong")
		}

		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
