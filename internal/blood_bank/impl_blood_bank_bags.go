package blood_bank

import (
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type implBloodBankBagsAPI struct {
	logger              zerolog.Logger
	tracer              trace.Tracer
	bagsCreatedCounter  metric.Int64Counter
	bagsUpdatedCounter  metric.Int64Counter
	bagsDeletedCounter  metric.Int64Counter
}

func NewBloodBankBagsApi() BloodBankBagsAPI {
	meter := otel.Meter("blood-bank")

	bagsCreatedCounter, err := meter.Int64Counter(
		"blood_bank_bags_created_total",
		metric.WithDescription("Total number of blood bags created"),
	)
	if err != nil {
		panic(err)
	}

	bagsUpdatedCounter, err := meter.Int64Counter(
		"blood_bank_bags_updated_total",
		metric.WithDescription("Total number of blood bags updated"),
	)
	if err != nil {
		panic(err)
	}

	bagsDeletedCounter, err := meter.Int64Counter(
		"blood_bank_bags_deleted_total",
		metric.WithDescription("Total number of blood bags deleted"),
	)
	if err != nil {
		panic(err)
	}

	return &implBloodBankBagsAPI{
		logger:             log.With().Str("component", "blood-bank").Logger(),
		tracer:             otel.Tracer("blood-bank"),
		bagsCreatedCounter: bagsCreatedCounter,
		bagsUpdatedCounter: bagsUpdatedCounter,
		bagsDeletedCounter: bagsDeletedCounter,
	}
}

var validBloodGroups = []string{"A", "B", "AB", "O"}
var validRhFactors = []string{"positive", "negative"}
var validStatuses = []string{"available", "reserved", "issued", "expired", "destroyed"}

// CreateBloodBag - Registers a new blood bag
func (o *implBloodBankBagsAPI) CreateBloodBag(c *gin.Context) {
	ctx, span := o.tracer.Start(c.Request.Context(), "CreateBloodBag")
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	updateBloodBankFunc(c, func(c *gin.Context, bloodBank *BloodBank) (*BloodBank, interface{}, int) {
		var bag BloodBag

		if err := c.ShouldBindJSON(&bag); err != nil {
			span.SetStatus(codes.Error, "Failed to bind JSON")
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		if !slices.Contains(validBloodGroups, bag.BloodGroup) {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid blood group, must be one of: A, B, AB, O",
			}, http.StatusBadRequest
		}

		if !slices.Contains(validRhFactors, bag.RhFactor) {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid Rh factor, must be one of: positive, negative",
			}, http.StatusBadRequest
		}

		if bag.Volume <= 0 {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Volume must be greater than 0",
			}, http.StatusBadRequest
		}

		if bag.Id == "" || bag.Id == "@new" {
			bag.Id = uuid.NewString()
		}

		bag.Status = "available"
		bag.CreatedAt = time.Now()

		if bag.CollectionDate.IsZero() {
			bag.CollectionDate = time.Now()
		}

		conflictIdx := slices.IndexFunc(bloodBank.BloodBags, func(existing BloodBag) bool {
			return bag.Id == existing.Id
		})

		if conflictIdx >= 0 {
			span.SetStatus(codes.Error, "Blood bag already exists")
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Blood bag with this ID already exists",
			}, http.StatusConflict
		}

		bloodBank.BloodBags = append(bloodBank.BloodBags, bag)

		o.bagsCreatedCounter.Add(
			c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("blood_bank_id", bloodBank.Id),
				attribute.String("blood_group", bag.BloodGroup),
			),
		)
		span.SetStatus(codes.Ok, "Blood bag created")
		return bloodBank, bag, http.StatusCreated
	})
}

// GetBloodBags - Provides the list of blood bags
func (o *implBloodBankBagsAPI) GetBloodBags(c *gin.Context) {
	updateBloodBankFunc(c, func(c *gin.Context, bloodBank *BloodBank) (*BloodBank, interface{}, int) {
		result := bloodBank.BloodBags
		if result == nil {
			result = []BloodBag{}
		}

		bloodGroup := c.Query("bloodGroup")
		rhFactor := c.Query("rhFactor")

		if bloodGroup != "" || rhFactor != "" {
			filtered := []BloodBag{}
			for _, bag := range result {
				if bloodGroup != "" && bag.BloodGroup != bloodGroup {
					continue
				}
				if rhFactor != "" && bag.RhFactor != rhFactor {
					continue
				}
				filtered = append(filtered, bag)
			}
			result = filtered
		}

		return nil, result, http.StatusOK
	})
}

// GetBloodBag - Provides details about a specific blood bag
func (o *implBloodBankBagsAPI) GetBloodBag(c *gin.Context) {
	updateBloodBankFunc(c, func(c *gin.Context, bloodBank *BloodBank) (*BloodBank, interface{}, int) {
		bagId := c.Param("bagId")

		if bagId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Bag ID is required",
			}, http.StatusBadRequest
		}

		bagIdx := slices.IndexFunc(bloodBank.BloodBags, func(bag BloodBag) bool {
			return bagId == bag.Id
		})

		if bagIdx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Blood bag not found",
			}, http.StatusNotFound
		}

		return nil, bloodBank.BloodBags[bagIdx], http.StatusOK
	})
}

// UpdateBloodBag - Updates specific blood bag
func (o *implBloodBankBagsAPI) UpdateBloodBag(c *gin.Context) {
	ctx, span := o.tracer.Start(c.Request.Context(), "UpdateBloodBag")
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	updateBloodBankFunc(c, func(c *gin.Context, bloodBank *BloodBank) (*BloodBank, interface{}, int) {
		var bag BloodBag

		if err := c.ShouldBindJSON(&bag); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		bagId := c.Param("bagId")

		if bagId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Bag ID is required",
			}, http.StatusBadRequest
		}

		bagIdx := slices.IndexFunc(bloodBank.BloodBags, func(existing BloodBag) bool {
			return bagId == existing.Id
		})

		if bagIdx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Blood bag not found",
			}, http.StatusNotFound
		}

		if bag.BloodGroup != "" {
			if !slices.Contains(validBloodGroups, bag.BloodGroup) {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid blood group",
				}, http.StatusBadRequest
			}
			bloodBank.BloodBags[bagIdx].BloodGroup = bag.BloodGroup
		}

		if bag.RhFactor != "" {
			if !slices.Contains(validRhFactors, bag.RhFactor) {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid Rh factor",
				}, http.StatusBadRequest
			}
			bloodBank.BloodBags[bagIdx].RhFactor = bag.RhFactor
		}

		if bag.Volume > 0 {
			bloodBank.BloodBags[bagIdx].Volume = bag.Volume
		}

		if bag.Status != "" {
			if !slices.Contains(validStatuses, bag.Status) {
				return nil, gin.H{
					"status":  http.StatusBadRequest,
					"message": "Invalid status",
				}, http.StatusBadRequest
			}
			bloodBank.BloodBags[bagIdx].Status = bag.Status
		}

		if bag.DonorId != "" {
			bloodBank.BloodBags[bagIdx].DonorId = bag.DonorId
		}

		if bag.Notes != "" {
			bloodBank.BloodBags[bagIdx].Notes = bag.Notes
		}

		if !bag.CollectionDate.IsZero() {
			bloodBank.BloodBags[bagIdx].CollectionDate = bag.CollectionDate
		}

		o.bagsUpdatedCounter.Add(
			c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("blood_bank_id", bloodBank.Id),
			),
		)
		return bloodBank, bloodBank.BloodBags[bagIdx], http.StatusOK
	})
}

// DeleteBloodBag - Deletes specific blood bag
func (o *implBloodBankBagsAPI) DeleteBloodBag(c *gin.Context) {
	updateBloodBankFunc(c, func(c *gin.Context, bloodBank *BloodBank) (*BloodBank, interface{}, int) {
		bagId := c.Param("bagId")

		if bagId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Bag ID is required",
			}, http.StatusBadRequest
		}

		bagIdx := slices.IndexFunc(bloodBank.BloodBags, func(bag BloodBag) bool {
			return bagId == bag.Id
		})

		if bagIdx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Blood bag not found",
			}, http.StatusNotFound
		}

		bloodBank.BloodBags = append(bloodBank.BloodBags[:bagIdx], bloodBank.BloodBags[bagIdx+1:]...)
		o.bagsDeletedCounter.Add(
			c.Request.Context(), 1,
			metric.WithAttributes(
				attribute.String("blood_bank_id", bloodBank.Id),
			),
		)
		return bloodBank, nil, http.StatusNoContent
	})
}
