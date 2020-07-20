package example

const (
	helloUserTKey = "Hello user"
)

var translations = map[string]map[string]string{
	"en": {
		helloUserTKey: "Hello user",
	},
	"fr": {
		helloUserTKey: "Bonjour utilisateur",
	},
	"nl": {
		helloUserTKey: "Hallo gebruiker",
	},
}

func TranslateFunc(lang string) func(string) string {
	tt, ok := translations[lang]
	if !ok {
		return func(s string) string {
			return s
		}
	}

	return func(s string) string {
		t, ok := tt[s]
		if !ok {
			return s
		}

		return t
	}
}

type Context struct {
	Page        Page
	User        User
	Leaderboard []User

	GetText func(string) string
}

func ExampleContext() Context {
	link := User{
		Name:    "Link",
		Email:   "link@hyrule.com",
		Logo:    "https://gamepedia.cursecdn.com/zelda_gamepedia_en/thumb/7/75/Young_Link_Navi.png/170px-Young_Link_Navi.png?version=e3a36be2fb9994d28b78ccb80754daf7",
		Premium: true,
		Points:  42,
	}

	return Context{
		Page: Page{
			Title: "The Game of the Year!",
			Lang:  "en",
		},
		User: link,
		Leaderboard: []User{
			{
				Name:    "Mario",
				Logo:    "https://vignette.wikia.nocookie.net/fantendo/images/7/77/BalloonMario.png/revision/latest/scale-to-width-down/185?cb=20100425175507",
				Premium: true,
				Points:  256,
			},
			{
				Name:    "<strong>Bowser",
				Logo:    "https://ubistatic19-a.akamaihd.net/ubicomstatic/en-GB/global/game-info/Luigi_150_289387.png",
				Premium: true,
				Points:  122,
			},
			{
				Name:   "Toad",
				Logo:   "https://vignette1.wikia.nocookie.net/fantendo/images/d/d5/Red_Blue_Toad.png/revision/latest/scale-to-width-down/150?cb=20110429150513",
				Points: 68,
			},
			link, // 42 points
			{
				Name:      "Wario",
				Logo:      "https://wiimedia.ign.com/wii/image/article/809/809868/Newcomer_Wario_1187977542.jpg",
				Anonymous: true,
				Premium:   true,
				Points:    12,
			},
		},
	}
}

type Page struct {
	Title string
	Lang  string
}

type User struct {
	Name      string
	Email     string // Email is only populated for the logged-in user
	Logo      string
	Anonymous bool
	Premium   bool
	Points    uint64
}

func (u User) ShowPremium() bool {
	return u.Premium && !u.Anonymous
}

func (u User) IsLoggedIn() bool {
	return u.Email != ""
}
