package blood_bank

import "github.com/gin-gonic/gin"

type BloodBankBagsAPI interface {
	GetBloodBags(c *gin.Context)
	CreateBloodBag(c *gin.Context)
	GetBloodBag(c *gin.Context)
	UpdateBloodBag(c *gin.Context)
	DeleteBloodBag(c *gin.Context)
}
