package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"wildberries-tech/internal/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveOrder(order models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockRepository) GetOrder(orderUID string) (*models.Order, error) {
	args := m.Called(orderUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockRepository) GetAllOrders() ([]models.Order, error) {
	args := m.Called()
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(orderUID string, order models.Order) {
	m.Called(orderUID, order)
}

func (m *MockCache) Get(orderUID string) (models.Order, bool) {
	args := m.Called(orderUID)
	return args.Get(0).(models.Order), args.Bool(1)
}

func (m *MockCache) LoadFromDB(orders []models.Order) {
	m.Called(orders)
}

func TestGetOrder_CacheHit(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	order := models.Order{OrderUID: "test-uid", TrackNumber: "TRACK123"}
	mockCache.On("Get", "test-uid").Return(order, true)

	h := New(mockRepo, mockCache)

	req, _ := http.NewRequest("GET", "/order/test-uid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/order/{order_uid}", h.GetOrder)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test-uid", response.OrderUID)

	mockCache.AssertExpectations(t)
}

func TestGetOrder_CacheMiss_DBHit(t *testing.T) {
	mockRepo := new(MockRepository)
	mockCache := new(MockCache)

	order := models.Order{OrderUID: "test-uid", TrackNumber: "TRACK123"}
	mockCache.On("Get", "test-uid").Return(models.Order{}, false)
	mockRepo.On("GetOrder", "test-uid").Return(&order, nil)
	mockCache.On("Set", "test-uid", order).Return()

	h := New(mockRepo, mockCache)

	req, _ := http.NewRequest("GET", "/order/test-uid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/order/{order_uid}", h.GetOrder)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
