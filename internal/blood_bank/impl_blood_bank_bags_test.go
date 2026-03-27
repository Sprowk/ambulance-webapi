package blood_bank

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/Sprowk/ambulance-webapi/internal/db_service"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace/noop"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
)

type BloodBankBagsSuite struct {
	suite.Suite
	dbServiceMock *DbServiceMock[BloodBank]
}

func TestBloodBankBagsSuite(t *testing.T) {
	suite.Run(t, new(BloodBankBagsSuite))
}

type DbServiceMock[DocType interface{}] struct {
	mock.Mock
}

func (this *DbServiceMock[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	args := this.Called(ctx, id, document)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	args := this.Called(ctx, id)
	return args.Get(0).(*DocType), args.Error(1)
}

func (this *DbServiceMock[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	args := this.Called(ctx, id, document)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) DeleteDocument(ctx context.Context, id string) error {
	args := this.Called(ctx, id)
	return args.Error(0)
}

func (this *DbServiceMock[DocType]) Disconnect(ctx context.Context) error {
	args := this.Called(ctx)
	return args.Error(0)
}

func (suite *BloodBankBagsSuite) SetupTest() {
	suite.dbServiceMock = &DbServiceMock[BloodBank]{}

	// Compile time Assert that the mock is of type db_service.DbService[BloodBank]
	var _ db_service.DbService[BloodBank] = suite.dbServiceMock

	suite.dbServiceMock.
		On("FindDocument", mock.Anything, mock.Anything).
		Return(
			&BloodBank{
				Id: "test-blood-bank",
				BloodBags: []BloodBag{
					{
						Id:             "test-bag",
						BloodGroup:     "A",
						RhFactor:       "positive",
						CollectionDate: time.Now(),
						Volume:         450,
						Status:         "available",
						DonorId:        "donor-001",
					},
				},
			},
			nil,
		)
}

func (suite *BloodBankBagsSuite) Test_UpdateBloodBag_DbServiceUpdateCalled() {
	// ARRANGE
	suite.dbServiceMock.
		On("UpdateDocument", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	json := `{
		"id": "test-bag",
		"bloodGroup": "B",
		"rhFactor": "negative",
		"volume": 350
	}`

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("db_service", suite.dbServiceMock)
	ctx.Params = []gin.Param{
		{Key: "bloodBankId", Value: "test-blood-bank"},
		{Key: "bagId", Value: "test-bag"},
	}
	ctx.Request = httptest.NewRequest("PUT", "/blood-bank/test-blood-bank/bags/test-bag", strings.NewReader(json))

	sut := implBloodBankBagsAPI{
		tracer:             noop.NewTracerProvider().Tracer("blood-bank"),
		logger:             zerolog.Nop(),
		bagsCreatedCounter: metricNoop.Int64Counter{},
		bagsUpdatedCounter: metricNoop.Int64Counter{},
		bagsDeletedCounter: metricNoop.Int64Counter{},
	}

	// ACT
	sut.UpdateBloodBag(ctx)

	// ASSERT
	suite.dbServiceMock.AssertCalled(suite.T(), "UpdateDocument", mock.Anything, "test-blood-bank", mock.Anything)
}
