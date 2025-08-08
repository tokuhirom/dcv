package models

type GenericContainer interface {
	IsDind() bool
	GetID() string
	GetName() string
}
