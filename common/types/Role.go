package types

type Role struct {
	Name        string              `json:"name"`
	Permissions map[string][]string `json:"permissions"`
}
