package handlers

import (
	s "project.com/myproject/stores"
)

type Handler struct {
	Store *s.PostgresStore
}

func NewHandler(store *s.PostgresStore) *Handler {
	return &Handler{Store: store}
}
