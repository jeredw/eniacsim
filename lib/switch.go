package lib

type Switch interface {
	Get() string
	Set(value string) error
}
