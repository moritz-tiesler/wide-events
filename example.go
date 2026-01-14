package wide_events

import (
	"context"
	"log/slog"
	"net/http"
)

func HandleCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add business context as it becomes available
	userID := "user_456"
	Record(ctx,
		slog.String("user.id", userID),
		slog.String("user.tier", "premium"),
		slog.Int("cart.item_count", 3),
	)

	err := processPayment(ctx)
	if err != nil {
		Record(ctx,
			slog.String("error.type", "payment_failed"),
			slog.String("error.msg", err.Error()),
		)
		http.Error(w, "Payment Failed", 500)
		return
	}

	Record(ctx, slog.String("payment.status", "success"))
	w.WriteHeader(http.StatusOK)
}

func processPayment(ctx context.Context) error {
	return nil
}

var handlerWithWideEvent = WideEventMiddleware(http.HandlerFunc(HandleCheckout))
