package kafka

import (
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
	consumer := NewConsumer(repo, cache, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	consumer.dlqProducer = dlqProducer

	validOrder := createValidOrder()
	validJSON, _ := json.Marshal(validOrder)

	// Expectation: Save to Repo, Set to Cache
	repo.On("SaveOrder", mock.AnythingOfType("models.Order")).Return(nil)
	cache.On("Set", validOrder.OrderUID, mock.AnythingOfType("models.Order")).Return()

	consumer.processMessage(validJSON)

	repo.AssertCalled(t, "SaveOrder", mock.AnythingOfType("models.Order"))
	cache.AssertCalled(t, "Set", validOrder.OrderUID, mock.AnythingOfType("models.Order"))
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	consumer := NewConsumer(repo, cache, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	invalidJSON := []byte(`{invalid-json}`)

	consumer.processMessage(invalidJSON)

	repo.AssertNotCalled(t, "SaveOrder")
	cache.AssertNotCalled(t, "Set")
}

func TestProcessMessage_ValidationFail(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	consumer := NewConsumer(repo, cache, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	invalidOrder := models.Order{OrderUID: ""}
	invalidJSON, _ := json.Marshal(invalidOrder)

	consumer.processMessage(invalidJSON)

	repo.AssertNotCalled(t, "SaveOrder")
	cache.AssertNotCalled(t, "Set")
}

func TestProcessMessage_RepoError(t *testing.T) {
	repo := new(MockRepo)
	cache := new(MockCache)
	consumer := NewConsumer(repo, cache, []string{"mock"}, "mock", "dlq-mock")

	dlqProducer := mocks.NewSyncProducer(t, nil)
	dlqProducer.ExpectSendMessageAndSucceed()
	consumer.dlqProducer = dlqProducer

	validOrder := createValidOrder()
	validJSON, _ := json.Marshal(validOrder)

	repo.On("SaveOrder", mock.Anything).Return(errors.New("db error"))

	consumer.processMessage(validJSON)

	repo.AssertCalled(t, "SaveOrder", mock.Anything)
	cache.AssertNotCalled(t, "Set")
}
