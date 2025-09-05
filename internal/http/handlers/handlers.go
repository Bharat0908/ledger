package handlers

// Package handlers provides HTTP handler functions for the ledger service.
// It includes routing, request parsing, and response formatting logic.
// The package leverages third-party libraries such as chi for routing and uuid for unique identifiers.
// It also interacts with internal packages like queue for background processing.
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

// AccountRepo defines the interface for account-related operations in the ledger system.
// It provides methods to create a new account and retrieve account information.
//
// Methods:
//   - CreateAccount: Creates a new account with the specified owner, currency, and initial balance.
//     Returns the UUID of the created account or an error if the operation fails.
//   - GetAccount: Retrieves the balance of the account identified by the given UUID.
//     Returns the account balance or an error if the account does not exist or retrieval fails.
type AccountRepo interface {
	CreateAccount(ctx context.Context, owner, currency string, initial int64) (uuid.UUID, error)
	GetAccount(ctx context.Context, id uuid.UUID) (int64, error)
}

// LedgerRepo defines the interface for accessing ledger transactions.
// It provides methods to retrieve transactions for a specific account.
type LedgerRepo interface {
	GetTransactions(ctx context.Context, accountID string, limit int) ([]map[string]interface{}, error)
}

// Handlers encapsulates dependencies required by HTTP handlers, including
// a message queue publisher, an account repository, and a ledger repository.
type Handlers struct {
	Pub        *queue.Publisher
	Repo       AccountRepo
	LedgerRepo LedgerRepo
}

// New creates and returns a new Handlers instance with the provided queue.Publisher,
// AccountRepo, and LedgerRepo. It initializes the Handlers struct with these dependencies
// for handling HTTP requests related to accounts and ledgers.
func New(pub *queue.Publisher, repo AccountRepo, lrepo LedgerRepo) *Handlers {
	return &Handlers{Pub: pub, Repo: repo, LedgerRepo: lrepo}
}

// Routes sets up and returns the HTTP routes for the ledger service, including endpoints for account creation,
// account retrieval, ledger retrieval, transaction and transfer enqueuing, as well as health and readiness checks.
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

// createAccount handles HTTP requests to create a new account.
// It expects a JSON payload with the account owner, currency, and initial balance.
// On success, it returns the created account's ID with HTTP status 201 Created.
// If the request body is invalid or an error occurs during account creation,
// it responds with the appropriate HTTP error code and message.
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

// getAccount handles HTTP requests to retrieve the balance of an account by its ID.
// It expects the account ID as a URL parameter, validates it, and fetches the account balance
// from the repository. If successful, it responds with a JSON object containing the balance.
// Returns a 400 error if the ID is invalid, or a 500 error if the repository operation fails.
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

// enqueueTx handles HTTP requests to enqueue a transaction message for processing.
// It expects a JSON payload containing account_id, type, amount, and an optional idempotency_key.
// If idempotency_key is not provided in the payload or headers, a new UUID is generated.
// The transaction message is published to the queue, and a response is returned with the status and idempotency key.
// Responds with 400 Bad Request on JSON decoding errors, 500 Internal Server Error on publishing failures,
// and 202 Accepted on successful queuing.
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

// enqueueTransfer handles HTTP requests to enqueue a money transfer operation.
// It expects a JSON payload containing the source account ID, destination account ID,
// transfer amount, and an optional idempotency key. If the idempotency key is not provided
// in the payload, it attempts to read it from the "Idempotency-Key" header, or generates
// a new UUID if none is found. The transfer request is published to a message queue for
// asynchronous processing. Responds with HTTP 202 Accepted and returns the idempotency key
// in the response body if successful, or an error message otherwise.
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

// getLedger handles HTTP requests to retrieve a limited number of ledger transactions for a given ledger ID.
// It extracts the "id" parameter from the URL, fetches up to 50 transactions from the LedgerRepo,
// and responds with a JSON object containing the entries. If an error occurs during retrieval,
// it responds with an HTTP 500 error and the error message.
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
