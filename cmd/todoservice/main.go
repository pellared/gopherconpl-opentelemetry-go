package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/XSAM/otelsql"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric/global"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/pellared/gopherconpl-opentelemetry-go/telemetry"
)

var db *sql.DB

const serviceName = "todoservice"

func main() {
	shutdownTracing, err := telemetry.SetupTracing(serviceName, "http://localhost:14268/api/traces")
	if err != nil {
		log.Fatalf("Failed to setup tracing: %v\n", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			log.Printf("Failed to shutdown tracing: %v\n", err)
		}
	}()

	shutdownMetrics, err := telemetry.SetupMetrics(serviceName)
	if err != nil {
		log.Fatalf("Failed to setup metrics: %v\n", err)
	}
	defer func() {
		if err := shutdownMetrics(context.Background()); err != nil {
			log.Printf("Failed to shutdown metrics: %v\n", err)
		}
	}()

	driverName, err := otelsql.Register("pgx", semconv.DBSystemPostgreSQL.Value.AsString())
	if err != nil {
		panic(err)
	}

	db, err = sql.Open(driverName, "postgres://postgres:pswd@localhost:5432/postgres")
	if err != nil {
		log.Fatalf("Unable to connection to database: %v\n", err)
	}

	r := mux.NewRouter()

	// Instrument gorilla/mux with OpenTelemetry tracing.
	r.Use(otelmux.Middleware("mux-server"))

	// Add a custom metric
	meter := global.Meter("todoservice")
	addTaskCnt, err := meter.NewInt64Counter("tasks_added")
	if err != nil {
		log.Fatalf("Unable to create list_tasks metrics counter: %v\n", err)
	}

	r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tasks, err := listTasks(ctx)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
		for _, t := range tasks {
			_, _ = fmt.Fprintf(rw, "%d. %s\n", t.id, t.description)
		}
	}).Methods("GET")

	r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer r.Body.Close()
		desc, err := io.ReadAll(r.Body)
		if err != nil {
			telemetry.AddErrorEvent(ctx, err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		err = addTask(ctx, string(desc))
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusCreated)
		addTaskCnt.Add(ctx, 1)
	}).Methods("POST")

	r.HandleFunc("/{id:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer r.Body.Close()
		desc, err := io.ReadAll(r.Body)
		if err != nil {
			telemetry.AddErrorEvent(ctx, err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			telemetry.AddErrorEvent(ctx, err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		err = updateTask(ctx, int32(id), string(desc))
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}).Methods("PUT")

	r.HandleFunc("/{id:[0-9]+}", func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			telemetry.AddErrorEvent(ctx, err)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		err = removeTask(ctx, int32(id))
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	// Instrument http.Handler with OpenTelemetry tracing and metrics.
	otelHandler := otelhttp.NewHandler(r, "http-server")

	log.Fatal(http.ListenAndServe(":8000", otelHandler))
}

type task struct {
	id          int32
	description string
}

func listTasks(ctx context.Context) ([]task, error) {
	var tasks []task
	rows, err := db.QueryContext(ctx, "SELECT id, description FROM tasks")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var t task
		if err := rows.Scan(&t.id, &t.description); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

func addTask(ctx context.Context, description string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO tasks(description) VALUES($1)", description)
	return err
}

func updateTask(ctx context.Context, itemNum int32, description string) error {
	_, err := db.ExecContext(ctx, "UPDATE tasks SET description=$1 WHERE id=$2", description, itemNum)
	return err
}

func removeTask(ctx context.Context, itemNum int32) error {
	_, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id=$1", itemNum)
	return err
}
