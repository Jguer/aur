package aur

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrServiceUnavailable represents a error when AUR is unavailable.
var ErrServiceUnavailable = errors.New("AUR is unavailable at this moment")

type PayloadError struct {
	StatusCode int
	ErrorField string
}

func (r *PayloadError) Error() string {
	return fmt.Sprintf("status %d: %s", r.StatusCode, r.ErrorField)
}

func GetErrorByStatusCode(code int) error {
	switch code {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	}

	return nil
}
