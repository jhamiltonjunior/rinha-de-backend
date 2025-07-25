package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/database"
	"github.com/jhamiltonjunior/rinha-de-backend/app/worker"
	"github.com/valyala/fasthttp"
)

type Details struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type TypeDetails struct {
	Default  Details `json:"default"`
	Fallback Details `json:"fallback"`
}

func Payments(ctx *fasthttp.RequestCtx) {
	bodyCopy := make([]byte, len(ctx.PostBody()))
	copy(bodyCopy, ctx.PostBody())

	cxt, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	paymentWorker := worker.PaymentWorker{
		Body:              bodyCopy,
		VouTeDarOContexto: cxt,
	}

	select {
	case worker.SegureOChann <- paymentWorker:
		sendJSONResponse(ctx, fasthttp.StatusAccepted)
	default:
		sendJSONResponse(ctx, fasthttp.StatusTooManyRequests)
	}
}

func PaymentsSummary(ctx *fasthttp.RequestCtx) {
	from := ctx.QueryArgs().Peek("from")
	to := ctx.QueryArgs().Peek("to")

	if len(from) == 0 {
		from = []byte("1970-01-01T00:00:00Z")
	}

	if len(to) == 0 {
		to = []byte("9999-12-31T23:59:59Z")
	}

	payments, err := database.GetPaymentHistory(database.MongoClient, string(from), string(to))
	if err != nil {
		fmt.Println("Erro ao buscar histórico de pagamentos:", err)
		sendJSONResponse(ctx, fasthttp.StatusInternalServerError)
		return
	}

	var typeDetails TypeDetails
	for _, payment := range payments {
		switch payment.Type {
		case "default":
			typeDetails.Default.TotalRequests++
			typeDetails.Default.TotalAmount += payment.Amount
		case "fallback":
			typeDetails.Fallback.TotalRequests++
			typeDetails.Fallback.TotalAmount += payment.Amount
		}
	}

	// ?from=2025-07-13T00:00:00&to=2025-07-13T14:33:48

	paymentsSummary, err := json.Marshal(typeDetails)
	if err != nil {
		fmt.Println("Erro ao serializar resumo de pagamentos:", err)
		sendJSONResponse(ctx, fasthttp.StatusInternalServerError)
		return
	}

	// fmt.Printf("?from=%s&to=%s\n", string(from), string(to))
	// fmt.Println("Payments Summary:", string(paymentsSummary))

	fmt.Fprintf(ctx, "%s", string(paymentsSummary))
	sendJSONResponse(ctx, fasthttp.StatusOK)
}

func PaymentsPurge(ctx *fasthttp.RequestCtx) {
	sendJSONResponse(ctx, fasthttp.StatusAccepted)
}
