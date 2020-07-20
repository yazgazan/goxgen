package views

import (
	gox "github.com/yazgazan/goxgen/src"
	exampletext "github.com/yazgazan/goxgen/src/example-text"
)

func Index(ctx exampletext.Context) gox.Writer {
	T := func(body gox.Writer) gox.Writer {
		s, err := body.String()
		if err != nil {
			return gox.Error(err)
		}

		return gox.Text(ctx.GetText(s))
	}

	return gox.NewComponent(&Main{Page: ctx.Page, User: ctx.User, Body: gox.Writers(gox.Text("\n"), gox.PlainText(T(gox.Writers(gox.Text("Hello user"))), gox.Text(" "), gox.Value(ctx.User.Name), gox.Text("!")), gox.Text("\n"), UserScore(ctx.User), gox.Text("\n"), gox.NewComponent(&LeaderBoard{Competitors: ctx.Leaderboard}), gox.Text("\n"), PlayButton(), gox.Text("\n"))})

}

func UserScore(u exampletext.User) gox.Writer {
	return gox.PlainText(gox.Text("You have earned "), gox.Value(u.Points), gox.Text(" !"))
}

func PlayButton() gox.Writer {
	return gox.PlainText(gox.Text("Play Now!"))
}
