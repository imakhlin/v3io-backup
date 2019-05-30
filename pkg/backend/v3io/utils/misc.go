package utils

import (
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"
	"net/http"
)

func IsNotExistsError(err error) bool {
	errorWithStatusCode, ok := err.(v3ioerrors.ErrorWithStatusCode)
	if !ok {
		// error of different type
		return false
	}
	// Ignore 404s
	if errorWithStatusCode.StatusCode() == http.StatusNotFound {
		return true
	}
	return false
}