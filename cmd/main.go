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

type AppState struct {
	Contacts Contacts
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
		Contacts: Contacts{},
	}
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", appState)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		contact := newContact(name, email)
		// TODO: Refactor it so that the response for form submission always updates the form itself and also returns the success message

		if appState.hasContact(&contact) {
			c.Response().Header().Add("HX-Retarget", "#error")
			c.Response().Header().Add("HX-Reswap", "innerHTML")
			return c.HTML(echo.ErrUnprocessableEntity.Code, "Contact with this email already exists")
		}

		appState.Contacts = append(appState.Contacts, contact)

		return c.Render(200, "contacts", appState)
	})

	e.Logger.Fatal(e.Start(":42069"))
}
