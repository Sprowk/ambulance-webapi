package blood_bank

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/Sprowk/ambulance-webapi/internal/db_service"
)

type implBloodBanksAPI struct {
}

func NewBloodBanksApi() BloodBanksAPI {
	return &implBloodBanksAPI{}
}

// CreateBloodBank - Creates new blood bank
func (o *implBloodBanksAPI) CreateBloodBank(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[BloodBank])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db context is not of required type",
				"error":   "cannot cast db context to db_service.DbService",
			})
		return
	}

	bloodBank := BloodBank{}
	err := c.BindJSON(&bloodBank)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		return
	}

	if bloodBank.Id == "" {
		bloodBank.Id = uuid.New().String()
	}

	err = db.CreateDocument(c, bloodBank.Id, &bloodBank)

	switch err {
	case nil:
		c.JSON(
			http.StatusCreated,
			bloodBank,
		)
	case db_service.ErrConflict:
		c.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Blood bank already exists",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create blood bank in database",
				"error":   err.Error(),
			},
		)
	}
}

func (o *implBloodBanksAPI) DeleteBloodBank(c *gin.Context) {
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[BloodBank])
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	bloodBankId := c.Param("bloodBankId")
	err := db.DeleteDocument(c, bloodBankId)

	switch err {
	case nil:
		c.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Blood bank not found",
				"error":   err.Error(),
			},
		)
	default:
		c.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete blood bank from database",
				"error":   err.Error(),
			})
	}
}
