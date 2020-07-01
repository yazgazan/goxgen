package views

import (
	"encoding/json"
	"fmt"

	gox "github.com/yazgazan/goxgen/src"
	"github.com/yazgazan/goxgen/src/example"
)

func Index(ctx example.Context) gox.HTML {
	T := func(body gox.HTML) gox.HTML {
		s, err := body.String()
		if err != nil {
			return gox.Error(err)
		}

		return gox.Text(ctx.GetText(s))
	}

	return gox.NewComponent(&Main{Page: ctx.Page, User: ctx.User, Body: gox.Writers(gox.Text("\n\t\t"), gox.Tag("script", gox.Markup(gox.Property("type", "text/javascript")), gox.Text("\n\t\t\tvar user = "), gox.Value(gox.Raw(JSON(ctx.User))), gox.Text(";\n\n\t\t\tconsole.log(user);\n\t\t")), gox.Text("\n\t\t"), gox.Tag("p", gox.Text("\n\t\t\t"), T(gox.Writers(gox.Text("Hello user"))), gox.Text(" "), gox.Value(ctx.User.Name), gox.Text("!\n\t\t")), gox.Text("\n\t\t"), UserScore(ctx.User, gox.Markup(gox.Property("style", "color: red;"))), gox.Text("\n\t\t"), gox.NewComponent(&LeaderBoard{Competitors: ctx.Leaderboard}), gox.Text("\n\t\t"), gox.Tag("div", gox.Markup(gox.Property("class", "play")), gox.Text("\n\t\t\t"), PlayButton(), gox.Text("\n\t\t")), gox.Text("\n\t"))})

}

func UserScore(u example.User, attrs gox.Attributes) gox.HTML {
	return gox.Tag("p", gox.Markup(attrs), gox.Text("\n\t\tYou have earned "), gox.Value(u.Points), gox.Text(" !\n    "))

}

func PlayButton() gox.HTML {
	return gox.Tag("button", gox.Text("Play Now!"))
}

func JSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err)
	}

	return string(b)
}
