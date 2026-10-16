package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
	"github.com/jcmexdev/flashcart-engine/services/order/internal/ports"
)

type OrderHandler struct {
	createOrderUseCase ports.CreateOrderUseCase
}

func NewOrderHandler(createOrderUseCase ports.CreateOrderUseCase) *OrderHandler {
	return &OrderHandler{
		createOrderUseCase: createOrderUseCase,
	}
}

// Routes registers all HTTP routes for order-service using the Chi router.
func (h *OrderHandler) Routes() http.Handler {
	r := chi.NewRouter()

	// Production-ready middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.CreateOrder)
	})

	return r
}

// CreateOrder handles the POST /orders request.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteUnifiedError(w, apperror.NewInvalidInput("malformed JSON body", map[string]string{
			"body": "could not parse request payload",
		}))
		return
	}

	output, err := h.createOrderUseCase.Execute(r.Context(), ports.CreateOrderInput{
		CustomerID:  req.CustomerID,
		TotalAmount: req.TotalAmount,
	})
	if err != nil {
		WriteUnifiedError(w, err)
		return
	}

	resp := OrderResponse{
		ID:          output.Order.ID,
		CustomerID:  output.Order.CustomerID,
		TotalAmount: output.Order.TotalAmount,
		Status:      output.Order.Status,
		CreatedAt:   output.Order.CreatedAt,
		UpdatedAt:   output.Order.UpdatedAt,
	}

	WriteJSON(w, http.StatusCreated, resp)
}
