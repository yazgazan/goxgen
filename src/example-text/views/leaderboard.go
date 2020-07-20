package views

import (
	gox "github.com/yazgazan/goxgen/src"
	exampletext "github.com/yazgazan/goxgen/src/example-text"
)

func LeaderBoardRow(u exampletext.User) gox.Writer {
	pStar := ""
	userName := u.Name

	if u.ShowPremium() {
		pStar = "* "
	}
	if u.Anonymous {
		userName = "John Doe"
	}

	return gox.PlainText(gox.Text("- "), gox.Value(pStar), gox.Value(userName), gox.Text(": "), gox.Value(u.Points))
}

type LeaderBoard struct {
	Competitors []exampletext.User
}

func (b LeaderBoard) Render() gox.Writer {
	rows := []gox.Writer{}

	for _, c := range b.Competitors {
		rows = append(rows, LeaderBoardRow(c))
	}

	return gox.PlainText(gox.Value(rows))
}
