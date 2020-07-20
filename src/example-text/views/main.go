package views

import (
	gox "github.com/yazgazan/goxgen/src"
	exampletext "github.com/yazgazan/goxgen/src/example-text"
)

type Main struct {
	Body gox.Writer
	Page exampletext.Page
	User exampletext.User
}

func (m Main) Render() gox.Writer {
	return gox.PlainText(gox.Text("\n"), gox.PlainText(gox.Value(m.Page.Title)), gox.Text("\n"), gox.NewComponent(&Header{User: m.User}), gox.Text("\n"), gox.Value(m.Body), gox.Text("\n"), Footer(), gox.Text("\n"))

}

type Header struct {
	User exampletext.User
}

func (h Header) Render() gox.Writer {
	loginButtonText := "Signup"

	if h.User.IsLoggedIn() {
		loginButtonText = "Profile"
	}

	return gox.PlainText(gox.Text("\n"), gox.PlainText(gox.Text("- Leaderboard")), gox.Text("\n"), gox.PlainText(gox.Text("- Forum")), gox.Text("\n"), gox.PlainText(gox.Text("- "), gox.Value(loginButtonText)), gox.Text("\n\n"), PlayButton(), gox.Text("\n"))

}

func Footer() gox.Writer {
	return gox.PlainText(gox.Text("Some legal mumbo jumbo."))
}
