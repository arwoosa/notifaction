package dao

type FetchIdentityResponse []*identity

type identity struct {
	Id     string `json:"id"`
	State  string `json:"state"`
	Traits struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Language string `json:"language"`
	} `json:"traits"`
}
