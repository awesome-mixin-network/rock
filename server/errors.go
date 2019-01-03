package server

import (
	"net/http"

	"github.com/soooooooon/gin-contrib/errors"
)

var (
	internalErr   = errors.New(1000, "server error, retry after a while", http.StatusInternalServerError)
	pageNotFound  = errors.New(404, "page not found", http.StatusNotFound)
	loginRequired = errors.New(1001, "login required", http.StatusUnauthorized)
)
