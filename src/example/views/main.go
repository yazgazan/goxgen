package views

import (
	gox "github.com/yazgazan/goxgen/src"
	"github.com/yazgazan/goxgen/src/example"
)

func MainStyle() gox.HTML {
	return gox.Tag("style", gox.Value(gox.Raw(`
		body>h1 {
			text-align: center;
		}

		#menu {
			width: 50%;
			margin: auto;
		}
		#menu>ul {
			text-align: center;
		}
		#menu>ul>li {
			display: inline-block;
			width: 10em;
			margin-left: 1em;
			margin-right: 1em;
			padding: .1em;
			background-color: lightgrey;
		}

		#header>button {
			float: right;
			margin-right: 2em;
			margin-top: -3em;
		}

		body>p {
			width: 75%;
			margin: auto;
			margin-top: 1em;
			margin-bottom: 1em;
		}

		table {
			width: 75%;
			margin: auto;
			border-collapse: collapse;
		}
		table td {
			border-bottom: 1px solid black;
		}
		table img {
			width: 3em;
		}

		.play {
			padding-top: 3em;
			text-align: center;
		}

		.legal {
			width: 50%;
			margin: auto;
			text-align: center;
			color: grey;
			margin-top: 3em;
			font-size: 0.7em;
		}

		.premium {
			color: orange;
		}
	`)))
}

type Main struct {
	Body gox.HTML
	Page example.Page
	User example.User
}

func (m Main) Render() gox.HTML {
	return gox.Tag("html", gox.Text("\n\t\t"), MainStyle(), gox.Text("\n\t\t"), gox.Tag("body", gox.Text("\n\t\t\t"), gox.Tag("h1", gox.Value(m.Page.Title)), gox.Text("\n\t\t\t"), gox.NewComponent(&Header{User: m.User}), gox.Text("\n\t\t\t"), gox.Value(m.Body), gox.Text("\n\t\t\t"), Footer(), gox.Text("\n\t\t")), gox.Text("\n\t"))

}

type Header struct {
	User example.User
}

func (h Header) Render() gox.HTML {
	loginButtonText := "Signup"

	if h.User.IsLoggedIn() {
		loginButtonText = "Profile"
	}

	return gox.Tag("div", gox.Markup(gox.Property("id", "header")), gox.Text("\n\t\t"), gox.Tag("div", gox.Markup(gox.Property("id", "menu")), gox.Text("\n\t\t\t"), gox.Tag("ul", gox.Text("\n\t\t\t\t"), gox.Tag("li", gox.Text("Leaderboard")), gox.Text("\n\t\t\t\t"), gox.Tag("li", gox.Text("Forum")), gox.Text("\n\t\t\t\t"), gox.Tag("li", gox.Value(loginButtonText)), gox.Text("\n\t\t\t")), gox.Text("\n\t\t")), gox.Text("\n\t\t"), PlayButton(), gox.Text("\n\t"))

}

func Footer() gox.HTML {
	return gox.Tag("p", gox.Markup(gox.Property("class", "legal")), gox.Text("\n\t\tSome legal mumbo jumbo.\n\t"))

}
