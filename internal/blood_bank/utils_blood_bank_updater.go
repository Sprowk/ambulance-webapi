package blood_bank

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Sprowk/ambulance-webapi/internal/db_service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type bloodBankUpdater = func(
	ctx *gin.Context,
	bloodBank *BloodBank,
) (updatedBloodBank *BloodBank, responseContent interface{}, status int)

func updateBloodBankFunc(ctx *gin.Context, updater bloodBankUpdater) {
	tracer := otel.Tracer("blood-bank")
	spanCtx, span := tracer.Start(ctx.Request.Context(), "updateBloodBankFunc")
	defer span.End()

	ctx.Request = ctx.Request.WithContext(spanCtx)

	value, exists := ctx.Get("db_service")
	if !exists {
		span.SetStatus(codes.Error, "db_service not found")
		ctx.JSON(
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
		span.SetStatus(codes.Error, "db_service context is not of type db_service.DbService")
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	bloodBankId := ctx.Param("bloodBankId")

	bloodBank, err := db.FindDocument(ctx.Request.Context(), bloodBankId)

	switch err {
	case nil:
		// continue
	case db_service.ErrNotFound:
		span.SetStatus(codes.Error, "Blood bank not found")
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Blood bank not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		span.SetStatus(codes.Error, "Failed to load blood bank from database")
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load blood bank from database",
				"error":   err.Error(),
			})
		return
	}

	updatedBloodBank, responseObject, status := updater(ctx, bloodBank)

	if updatedBloodBank != nil {
		err = db.UpdateDocument(ctx.Request.Context(), bloodBankId, updatedBloodBank)
	} else {
		err = nil // redundant but for clarity
	}

	switch err {
	case nil:
		span.SetStatus(codes.Ok, "Blood bank updated")
		if responseObject != nil {
			ctx.JSON(status, responseObject)
		} else {
			ctx.AbortWithStatus(status)
		}
	case db_service.ErrNotFound:
		span.SetStatus(codes.Error, "Blood bank not found")
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Blood bank was deleted while processing the request",
				"error":   err.Error(),
			},
		)
	default:
		span.SetStatus(codes.Error, "Failed to update blood bank in database")
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update blood bank in database",
				"error":   err.Error(),
			})
	}
}
