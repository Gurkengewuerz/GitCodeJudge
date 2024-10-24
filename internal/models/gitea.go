package models

type GiteaPushEvent struct {
	Ref        string `json:"ref"`    // refs/heads/develop
	Before     string `json:"before"` // 28e1879d029cb852e4844d9c718537df08844e03
	After      string `json:"after"`  // bffeb74224043ba2feb48d137756c8a9331c449a
	Repository struct {
		Name     string `json:"name"`      // webhooks
		FullName string `json:"full_name"` // gitea/webhooks
		CloneURL string `json:"clone_url"` // http://localhost:3000/gitea/webhooks.git
	} `json:"repository"`
}
