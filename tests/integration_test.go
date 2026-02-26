package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wildberries-tech/internal/cache"
	"wildberries-tech/internal/handlers"
	"wildberries-tech/internal/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository for Integration
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
	return nil
}

func TestIntegration_GetOrder(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	// Use real cache
	realCache := cache.New(5*time.Minute, 10*time.Minute)
	
	h := handlers.New(mockRepo, realCache)
	router := mux.NewRouter()
	router.HandleFunc("/order/{order_uid}", h.GetOrder)

	// 1. Test Cache Miss -> DB Hit -> Cache Fill
	orderUID := "test-integration-uid"
	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "TRACK_INT_1",
		// Minimal valid fields
	}
	
	mockRepo.On("GetOrder", orderUID).Return(&order, nil)

	req, _ := http.NewRequest("GET", "/order/"+orderUID, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respOrder models.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &respOrder); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, orderUID, respOrder.OrderUID)

	// Verify it's in cache now
	cachedOrder, found := realCache.Get(orderUID)
	assert.True(t, found)
	assert.Equal(t, orderUID, cachedOrder.OrderUID)

	// 2. Test Cache Hit (Repo not called)
	// Reset mocks to ensure GetOrder is NOT called again
	mockRepo.Calls = nil // Clear calls
	mockRepo.ExpectedCalls = nil
	// We don't expect GetOrder call because it's in cache

	req2, _ := http.NewRequest("GET", "/order/"+orderUID, nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	
	assert.Equal(t, http.StatusOK, rr2.Code)
}