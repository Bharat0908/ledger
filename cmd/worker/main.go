package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Bharat0908/ledger/internal/queue"
	"github.com/Bharat0908/ledger/internal/repo"
)

// main is the entry point for the worker service. It initializes connections to PostgreSQL (via pgxpool),
// MongoDB, and RabbitMQ using environment variables for configuration. The function sets up repositories
// for both databases, constructs a transaction applier and a ledger writer, and starts a queue consumer
// to process incoming messages. It listens for system interrupt or termination signals to gracefully
// shut down the worker, allowing time for cleanup before exiting.
func main() {
	ctx := context.Background()
	pg, err := pgxpool.New(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatalf("pgxpool: %v", err)
	}
	defer pg.Close()

	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	defer mc.Disconnect(ctx)

	mcol := mc.Database("ledger").Collection("entries")

	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("amqp dial: %v", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("amqp channel: %v", err)
	}
	defer ch.Close()

	pgRepo := &repo.PGRepo{DB: pg}
	mongoRepo := &repo.MongoRepo{C: mcol}

	txApplier := &workerApplier{pg: pgRepo}
	ledgerWriter := &workerLedgerWriter{m: mongoRepo}

	consumer := &queue.Consumer{Ch: ch, Queue: "tx-queue", Applier: txApplier, Ledger: ledgerWriter}

	// start consumer
	go func() {
		if err := consumer.Start(context.Background()); err != nil {
			log.Fatalf("consumer error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("worker shutting down")
	// allow some time for cleanup
	time.Sleep(2 * time.Second)
}

// small adapters
type workerApplier struct{ pg *repo.PGRepo }

func (w *workerApplier) Apply(ctx context.Context, accID, typ string, amount int64, key string) (int64, error) {
	id, err := uuid.Parse(accID)
	if err != nil {
		return 0, err
	}
	return w.pg.ApplyTransaction(ctx, id, typ, amount, key)
}

func (w *workerApplier) ApplyTransfer(ctx context.Context, from, to string, amount int64, key string) (int64, int64, error) {
	fid, err := uuid.Parse(from)
	if err != nil {
		return 0, 0, err
	}
	tid, err := uuid.Parse(to)
	if err != nil {
		return 0, 0, err
	}
	return w.pg.ApplyTransfer(ctx, fid, tid, amount, key)
}

type workerLedgerWriter struct{ m *repo.MongoRepo }

func (w *workerLedgerWriter) Write(ctx context.Context, accID, typ string, amount, balanceAfter int64, key string, at time.Time) error {
	id, err := uuid.Parse(accID)
	if err != nil {
		return err
	}
	return w.m.InsertLedger(ctx, id, typ, amount, balanceAfter, key, at)
}

func (w *workerLedgerWriter) WriteTransfer(ctx context.Context, from, to string, amount, fromAfter, toAfter int64, key string, at time.Time) error {
	fid, err := uuid.Parse(from)
	if err != nil {
		return err
	}
	tid, err := uuid.Parse(to)
	if err != nil {
		return err
	}
	return w.m.InsertTransferLedger(ctx, fid, tid, amount, fromAfter, toAfter, key, at)
}
