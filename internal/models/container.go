package models

type Container struct {
	Name      string
	Image     string
	ID        string
	Status    string
	State     string
	Service   string
	CreatedAt string
	Ports     string
}

type Process struct {
	Container
	IsDind bool
}