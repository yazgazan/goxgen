package views

import (
	"encoding/json"
	"fmt"

	"github.com/yazgazan/goxgen/src/example"
	"github.com/yazgazan/goxgen/src"
)

func Index(ctx example.Context) gox.ComponentOrHTML {
	T := func(body gox.ComponentOrHTML) gox.ComponentOrHTML {
		return gox.Text(ctx.GetText(body.Render()))
	}

	return gox.NewComponent(&Main{Page: ctx.Page, User: ctx.User, Body: gox.Text("", "\n\t\t",
		gox.Tag("script", gox.Markup(gox.Property("type", "text/javascript")), gox.Text("\n\t\t\tvar user = "), gox.Value(gox.Raw(JSON(ctx.User))), gox.Text(";\n\n\t\t\tconsole.log(user);\n\t\t")),

		"\n\t\t",
		gox.Tag("p", gox.Text("\n\t\t\t"), T(gox.Text("", "Hello user")), gox.Text(" "), gox.Value(ctx.User.Name), gox.Text("!\n\t\t")),
		"\n\t\t",
		UserScore(ctx.User, gox.Markup(gox.Property("style", "color: red;"))), "\n\t\t",
		gox.NewComponent(&LeaderBoard{Competitors: ctx.Leaderboard}), "\n\t\t",
		gox.Tag("div", gox.Markup(gox.Property("class", "play")), gox.Text("\n\t\t\t"), PlayButton(), gox.Text("\n\t\t")),
		"\n\t")})

}

func UserScore(u example.User, attrs gox.Applyer) gox.ComponentOrHTML {
	return gox.Tag("p", gox.Markup(attrs), gox.Text("\n\t\tYou have earned "), gox.Value(u.Points), gox.Text(" !\n    "))

}

func PlayButton() gox.ComponentOrHTML {
	return gox.Tag("button", gox.Text("Play Now!"))
}

func JSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err)
	}

	return string(b)
}
