package transaction

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TxController struct {
	service *TxService
	logger  *log.Logger
}

func NewTxController(service *TxService, logger *log.Logger) *TxController {
	return &TxController{
		service: service,
		logger:  logger,
	}
}

func (txController *TxController) GetTransactionByID(ctx *gin.Context) {
	id := ctx.Param("id")
	txID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction id"})
		return
	}

	reqCtx := ctx.Request.Context()
	reqCtx, cancel := context.WithTimeout(reqCtx, 5*time.Second)
	defer cancel()

	tx, err := txController.service.GetTransactionByID(reqCtx, txID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tx)
}

func (txController *TxController) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/transactions")
	{
		group.GET("/:id", txController.GetTransactionByID)
	}
}
