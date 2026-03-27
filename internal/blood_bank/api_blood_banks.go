package blood_bank

import "github.com/gin-gonic/gin"

type BloodBanksAPI interface {
	CreateBloodBank(c *gin.Context)
	DeleteBloodBank(c *gin.Context)
}
