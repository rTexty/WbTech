package kafka

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"wildberries-tech/internal/models"

	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) SaveOrder(order models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockRepo) GetOrder(uid string) (*models.Order, error) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockRepo) GetAllOrders() ([]models.Order, error) {
	args := m.Called()
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockRepo) Close() error {
	return nil
}

func (m *MockRepo) DB() (*sql.DB, error) {
	return nil, nil
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(uid string, order models.Order) {
	m.Called(uid, order)
}

func (m *MockCache) Get(uid string) (models.Order, bool) {
	args := m.Called(uid)
	return args.Get(0).(models.Order), args.Bool(1)
}

func (m *MockCache) LoadFromDB(orders []models.Order) {
	m.Called(orders)
}

// MockMetrics is a mock implementation of metrics.Metrics.
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncMessagesTotal(status string) {
	m.Called(status)
}

func (m *MockMetrics) SetResourceUp(resource string, up float64) {
	m.Called(resource, up)
}

func (m *MockMetrics) IncHTTPRequests(method, path, status string) {
	m.Called(method, path, status)
}

func (m *MockMetrics) ObserveHTTPDuration(method, path string, seconds float64) {
	m.Called(method, path, seconds)
}

// Helper to create a fully valid order
func createValidOrder() models.Order {
	return models.Order{
		OrderUID:    "valid-uid",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test User",
			Phone:   "+1234567890",
			Zip:     "123456",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@example.com",
		},
		Payment: models.Payment{
			Transaction:  "trans-123",
			RequestID:    "req-1",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       100,
			PaymentDt:    1620000000,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   100,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      123456,
				TrackNumber: "TRACK123",
				Price:       100,
				Rid:         "rid-1",
				Name:        "Item 1",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        1234567,
				Brand:       "TestBrand",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust-1",
		DeliveryService: "meest",
		Shardkey:        "9",
		SmID:            99,
		DateCreated:     time.Now(),
		OofShard:        "1",
	}
}

func TestProcessMessage(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	metricsM := new(MockMetrics)
	consumer := NewConsumer(repo, cache, metricsM, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	consumer.dlqProducer = dlqProducer

	validOrder := createValidOrder()
	validJSON, _ := json.Marshal(validOrder)

	// Expectation: Save to Repo, Set to Cache, increment metrics
	repo.On("SaveOrder", mock.AnythingOfType("models.Order")).Return(nil)
	cache.On("Set", validOrder.OrderUID, mock.AnythingOfType("models.Order")).Return()
	metricsM.On("IncMessagesTotal", "success").Return()

	consumer.processMessage(validJSON)

	repo.AssertCalled(t, "SaveOrder", mock.AnythingOfType("models.Order"))
	cache.AssertCalled(t, "Set", validOrder.OrderUID, mock.AnythingOfType("models.Order"))
	metricsM.AssertCalled(t, "IncMessagesTotal", "success")
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	metricsM := new(MockMetrics)
	consumer := NewConsumer(repo, cache, metricsM, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	metricsM.On("IncMessagesTotal", "error").Return()

	invalidJSON := []byte(`{invalid-json}`)

	consumer.processMessage(invalidJSON)

	repo.AssertNotCalled(t, "SaveOrder")
	cache.AssertNotCalled(t, "Set")
	metricsM.AssertCalled(t, "IncMessagesTotal", "error")
}

func TestProcessMessage_ValidationFail(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	metricsM := new(MockMetrics)
	consumer := NewConsumer(repo, cache, metricsM, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	metricsM.On("IncMessagesTotal", "error").Return()

	invalidOrder := models.Order{OrderUID: ""}
	invalidJSON, _ := json.Marshal(invalidOrder)

	consumer.processMessage(invalidJSON)

	repo.AssertNotCalled(t, "SaveOrder")
	cache.AssertNotCalled(t, "Set")
	metricsM.AssertCalled(t, "IncMessagesTotal", "error")
}

func TestProcessMessage_RepoError(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	metricsM := new(MockMetrics)
	consumer := NewConsumer(repo, cache, metricsM, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	validOrder := createValidOrder()
	validJSON, _ := json.Marshal(validOrder)

	repo.On("SaveOrder", mock.Anything).Return(errors.New("db error"))
	metricsM.On("IncMessagesTotal", "error").Return()

	consumer.processMessage(validJSON)

	repo.AssertCalled(t, "SaveOrder", mock.Anything)
	cache.AssertNotCalled(t, "Set")
	metricsM.AssertCalled(t, "IncMessagesTotal", "error")
}
