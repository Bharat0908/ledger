package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    amqp "github.com/rabbitmq/amqp091-go"

    handlers "github.com/Bharat0908/ledger/internal/http/handlers"
    "github.com/Bharat0908/ledger/internal/queue"
    "github.com/Bharat0908/ledger/internal/repo"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    ctx := context.Background()

    // Postgres
    pgDSN := os.Getenv("POSTGRES_DSN")
    if pgDSN == "" { pgDSN = "postgres://postgres:postgres@postgres:5432/ledger" }
    pg, err := pgxpool.New(ctx, pgDSN)
    if err != nil { log.Fatalf("pgxpool.New: %v", err) }
    defer pg.Close()


     // Mongo
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" { mongoURI = "mongodb://mongo:27017" }
    mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil { log.Fatalf("mongo connect: %v", err) }
    defer mc.Disconnect(ctx)
    mcol := mc.Database("ledger").Collection("entries")

    // RabbitMQ
    ramqpURL := os.Getenv("RABBITMQ_URL")
    if ramqpURL == "" { ramqpURL = "amqp://guest:guest@rabbitmq:5672/" }
    conn, err := amqp.Dial(ramqpURL)
    if err != nil { log.Fatalf("amqp dial: %v", err) }
    defer conn.Close()
    ch, err := conn.Channel()
    if err != nil { log.Fatalf("amqp channel: %v", err) }
    defer ch.Close()
    _ = ch.ExchangeDeclare("tx", "direct", true, false, false, false, nil)
    _, _ = ch.QueueDeclare("tx-queue", true, false, false, false, nil)
    _ = ch.QueueBind("tx-queue", "tx", "tx", false, nil)

    pub := queue.NewPublisher(ch, "tx", "tx")
    rep := &repo.PGRepo{DB: pg}
    mongoRepo := &repo.MongoRepo{C: mcol}

    h := handlers.New(pub, rep, mongoRepo)
    r := chi.NewRouter()
    r.Mount("/", h.Routes())

    srv := &http.Server{ Addr: ":8080", Handler: r, ReadTimeout: 10*time.Second, WriteTimeout: 10*time.Second }

    go func() {
        log.Println("api listening on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe: %v", err)
        }
    }()

    // graceful shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop
    log.Println("shutting down server...")
    ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctxShut); err != nil { log.Fatalf("shutdown error: %v", err) }
}
