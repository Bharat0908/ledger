package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Bharat0908/ledger/internal/queue"
	//"github.com/Bharat0908/ledger/internal/repo"
)

type AccountRepo interface {
	CreateAccount(ctx context.Context, owner, currency string, initial int64) (uuid.UUID, error)
	GetAccount(ctx context.Context, id uuid.UUID) (int64, error)
}

type LedgerRepo interface {
	GetTransactions(ctx context.Context, accountID string, limit int) ([]map[string]interface{}, error)
}

type Handlers struct {
	Pub        *queue.Publisher
	Repo       AccountRepo
	LedgerRepo LedgerRepo
}

func New(pub *queue.Publisher, repo AccountRepo, lrepo LedgerRepo) *Handlers {
	return &Handlers{Pub: pub, Repo: repo, LedgerRepo: lrepo}
}

func (h *Handlers) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/v1/accounts", h.createAccount)
	r.Get("/v1/accounts/{id}", h.getAccount)
	r.Get("/v1/accounts/{id}/ledger", h.getLedger)
	r.Post("/v1/transactions", h.enqueueTx)
	r.Post("/v1/transfers", h.enqueueTransfer)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
	return r
}

func (h *Handlers) createAccount(w http.ResponseWriter, r *http.Request) {
	type req struct {
		Owner, Currency string `json:"owner"`
		InitialBalance  int64  `json:"initial_balance"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	id, err := h.Repo.CreateAccount(r.Context(), body.Owner, body.Currency, body.InitialBalance)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

func (h *Handlers) getAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}
	bal, err := h.Repo.GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]int64{"balance": bal})
}

func (h *Handlers) enqueueTx(w http.ResponseWriter, r *http.Request) {
	type req struct {
		AccountID      string `json:"account_id"`
		Type           string `json:"type"`
		Amount         int64  `json:"amount"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	key := body.IdempotencyKey
	if key == "" {
		key = r.Header.Get("Idempotency-Key")
	}
	if key == "" {
		key = uuid.NewString()
	}
	msg := queue.TxMessage{AccountID: body.AccountID, Type: body.Type, Amount: body.Amount, Key: key, CreatedAt: time.Now()}
	if err := h.Pub.Publish(r.Context(), msg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued", "idempotency_key": key})
}

func (h *Handlers) enqueueTransfer(w http.ResponseWriter, r *http.Request) {
	type req struct {
		FromAccountID  string `json:"from_account_id"`
		ToAccountID    string `json:"to_account_id"`
		Amount         int64  `json:"amount"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	var body req
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	key := body.IdempotencyKey
	if key == "" {
		key = r.Header.Get("Idempotency-Key")
	}
	if key == "" {
		key = uuid.NewString()
	}
	msg := queue.TransferMessage{FromAccountID: body.FromAccountID, ToAccountID: body.ToAccountID, Amount: body.Amount, Key: key, CreatedAt: time.Now()}
	if err := h.Pub.PublishTransfer(r.Context(), msg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued", "idempotency_key": key})
}

func (h *Handlers) getLedger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit := 50
	entries, err := h.LedgerRepo.GetTransactions(r.Context(), id, limit)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": entries})
}
