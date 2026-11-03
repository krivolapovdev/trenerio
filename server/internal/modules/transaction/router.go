package transaction

import "github.com/gin-gonic/gin"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	txGroup := rg.Group("/transactions")
	{
		txGroup.GET("", r.handler.GetPaginatedTransactions)
		txGroup.POST("", r.handler.CreateTransaction)
		txGroup.PATCH("/:transactionId", r.handler.UpdateTransaction)
		txGroup.DELETE("/:transactionId", r.handler.DeleteTransaction)
	}
}
