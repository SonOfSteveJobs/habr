package middleware

type tokenError string

func (e tokenError) Error() string { return string(e) }

const (
	errInvalidToken tokenError = "invalid token"
	errTokenExpired tokenError = "token expired"
)
