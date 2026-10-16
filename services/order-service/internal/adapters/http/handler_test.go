package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httpadapter "github.com/jcmexdev/flashcart-engine/services/order/internal/adapters/http"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/domain"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/mocks"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	mockUC := mocks.NewMockCreateOrderUseCase(t)
	orderID := uuid.New()

	mockUC.EXPECT().
		Execute(mock.Anything, ports.CreateOrderInput{
			CustomerID:  "cust_999",
			TotalAmount: 150.50,
		}).
		Return(&ports.CreateOrderOutput{
			Order: &domain.Order{
				ID:          orderID,
				CustomerID:  "cust_999",
				TotalAmount: 150.50,
				Status:      domain.OrderStatusPending,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			},
		}, nil).
		Once()

	handler := httpadapter.NewOrderHandler(mockUC)
	router := handler.Routes()

	body := []byte(`{"customer_id":"cust_999","total_amount":150.50}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp httpadapter.OrderResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, orderID, resp.ID)
	assert.Equal(t, "cust_999", resp.CustomerID)
	assert.Equal(t, domain.OrderStatusPending, resp.Status)
}

func TestOrderHandler_CreateOrder_UnifiedError(t *testing.T) {
	mockUC := mocks.NewMockCreateOrderUseCase(t)

	mockUC.EXPECT().
		Execute(mock.Anything, ports.CreateOrderInput{
			CustomerID:  "",
			TotalAmount: 100,
		}).
		Return(nil, apperror.NewInvalidInput("invalid order parameters", map[string]string{
			"customer_id": "customer_id is required",
		})).
		Once()

	handler := httpadapter.NewOrderHandler(mockUC)
	router := handler.Routes()

	body := []byte(`{"customer_id":"","total_amount":100}`)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	var errResp httpadapter.UnifiedErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, apperror.CodeInvalidInput, errResp.Code)
	assert.Equal(t, "customer_id is required", errResp.Details["customer_id"])
}
