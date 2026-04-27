package github

type GHCloneURLTestCase struct {
	Name     string
	WebURL   string
	FullName string
	Token    string
	WantURL  string
}

func GHCloneURLTestCases() []GHCloneURLTestCase {
	return []GHCloneURLTestCase{
		{
			Name:     "github.com clone URL",
			WebURL:   "https://github.com",
			FullName: "user/repo",
			Token:    "ghp_token",
			WantURL:  "https://github.com/user/repo.git",
		},
		{
			Name:     "custom github enterprise clone URL",
			WebURL:   "https://github.mycompany.com",
			FullName: "user/repo",
			Token:    "token",
			WantURL:  "https://github.mycompany.com/user/repo.git",
		},
	}
}

type GHConfigureTestCase struct {
	Name    string
	Token   string
	APIURL  string
	WebURL  string
	WantErr bool
}

func GHConfigureTestCases() []GHConfigureTestCase {
	return []GHConfigureTestCase{
		{
			Name:    "valid configuration",
			Token:   "ghp_token",
			APIURL:  "https://api.github.com",
			WebURL:  "https://github.com",
			WantErr: false,
		},
		{
			Name:    "missing web URL",
			Token:   "ghp_token",
			APIURL:  "https://api.github.com",
			WebURL:  "",
			WantErr: true,
		},
	}
}