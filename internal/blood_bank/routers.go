package blood_bank

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name		string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method		string
	// Pattern is the pattern of the URI.
	Pattern	 	string
	// HandlerFunc is the handler function of this route.
	HandlerFunc	gin.HandlerFunc
}

// NewRouter returns a new router.
func NewRouter(handleFunctions ApiHandleFunctions) *gin.Engine {
	return NewRouterWithGinEngine(gin.Default(), handleFunctions)
}

// NewRouter add routes to existing gin engine.
func NewRouterWithGinEngine(router *gin.Engine, handleFunctions ApiHandleFunctions) *gin.Engine {
	for _, route := range getRoutes(handleFunctions) {
		if route.HandlerFunc == nil {
			route.HandlerFunc = DefaultHandleFunc
		}
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// Default handler for not yet implemented routes
func DefaultHandleFunc(c *gin.Context) {
	c.String(http.StatusNotImplemented, "501 not implemented")
}

type ApiHandleFunctions struct {
	// Routes for the BloodBankBagsAPI part of the API
	BloodBankBagsAPI BloodBankBagsAPI
	// Routes for the BloodBanksAPI part of the API
	BloodBanksAPI BloodBanksAPI
}

func getRoutes(handleFunctions ApiHandleFunctions) []Route {
	return []Route{
		{
			"CreateBloodBank",
			http.MethodPost,
			"/api/blood-bank",
			handleFunctions.BloodBanksAPI.CreateBloodBank,
		},
		{
			"DeleteBloodBank",
			http.MethodDelete,
			"/api/blood-bank/:bloodBankId",
			handleFunctions.BloodBanksAPI.DeleteBloodBank,
		},
		{
			"GetBloodBags",
			http.MethodGet,
			"/api/blood-bank/:bloodBankId/bags",
			handleFunctions.BloodBankBagsAPI.GetBloodBags,
		},
		{
			"CreateBloodBag",
			http.MethodPost,
			"/api/blood-bank/:bloodBankId/bags",
			handleFunctions.BloodBankBagsAPI.CreateBloodBag,
		},
		{
			"GetBloodBag",
			http.MethodGet,
			"/api/blood-bank/:bloodBankId/bags/:bagId",
			handleFunctions.BloodBankBagsAPI.GetBloodBag,
		},
		{
			"UpdateBloodBag",
			http.MethodPut,
			"/api/blood-bank/:bloodBankId/bags/:bagId",
			handleFunctions.BloodBankBagsAPI.UpdateBloodBag,
		},
		{
			"DeleteBloodBag",
			http.MethodDelete,
			"/api/blood-bank/:bloodBankId/bags/:bagId",
			handleFunctions.BloodBankBagsAPI.DeleteBloodBag,
		},
	}
}
