package views

import (
	gox "github.com/yazgazan/goxgen/src"
	"github.com/yazgazan/goxgen/src/example"
)

func LeaderBoardRow(u example.User) gox.HTML {
	rowClass := ""
	userName := u.Name

	if u.ShowPremium() {
		rowClass = "premium"
	}
	if u.Anonymous {
		userName = "John Doe"
	}

	return gox.Tag("tr", gox.Markup(gox.Property("class", rowClass)), gox.Text("\n\t\t"), gox.Tag("td", gox.Text("\n\t\t\t"), gox.Tag("img", gox.Markup(gox.Property("src", u.Logo))), gox.Text("\n\t\t")), gox.Text("\n\t\t"), gox.Tag("td", gox.Text("\n\t\t\t"), gox.Tag("p", gox.Value(userName)), gox.Text("\n\t\t")), gox.Text("\n\t\t"), gox.Tag("td", gox.Text("\n\t\t\t"), gox.Value(u.Points), gox.Text("\n\t\t")), gox.Text("\n\t"))

}

type LeaderBoard struct {
	Competitors []example.User
}

func (b LeaderBoard) Render() gox.HTML {
	rows := []gox.HTML{}

	for _, c := range b.Competitors {
		rows = append(rows, LeaderBoardRow(c))
	}

	return gox.Tag("table", gox.Text("\n\t\t"), gox.Value(rows), gox.Text("\n\t"))

}
