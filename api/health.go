package api

import (
	"net/http"
	"sync"
)

type (
	HealthDto struct {
		Status string `json:"status" binding:"required"`
	}

	healthHandler struct {
	}

	RestHeartBeat interface {
		Health() func(http.ResponseWriter, *http.Request)
	}
)

var onceHealthHandler sync.Once
var instanceHealthHandler *healthHandler

func HealthHandlerInstance() RestHeartBeat {
	onceHealthHandler.Do(func() {
		instanceHealthHandler = &healthHandler{}
	})
	return instanceHealthHandler
}

func (h *healthHandler) Health() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement a proper health check
		RespondWithJSON(w, http.StatusOK, &HealthDto{Status: "OK"})
	}
}
