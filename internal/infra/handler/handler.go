package handler

import "github.com/gin-gonic/gin"

type handler interface {
	Register(r *gin.Engine)
}

type Root struct {
	router  *gin.Engine
	handler []handler
}

func NewRoot(router *gin.Engine, handlers ...handler) *Root {
	return &Root{
		router:  router,
		handler: handlers,
	}
}

func (h *Root) RegisterRoutes() {
	for _, handler := range h.handler {
		handler.Register(h.router)
	}
}
