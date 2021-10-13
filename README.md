# Observability via OpenTelemtry Go - Quick Start

[Presentation](https://docs.google.com/presentation/d/1ir9eyBLxO4n3zPcOPhxJkZJ9cXJSo7KUhuRAadG7z8Q/edit?usp=sharing)

Run the distributed tracing backend:

```sh
docker run -d --name jaeger -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:1.26
```

Run the metrics backend:

```sh
docker run -d --name prometheus -p 9090:9090 -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus:v2.29.2
```

Run the database:

```sh
docker run -d --name todo-db -p 5432:5432 -e POSTGRES_PASSWORD=pswd -v $(pwd)/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql postgres:13-alpine
```

Build and run the service:

```sh
cd cmd/todoservice && go install && cd -
todoservice
```

Build and use the CLI app:

```sh
cd cmd/todo && go install && cd -
todo add "important work"
todo list
todo
todo add "very long description that is extremely important"
```

Navigate to <http://localhost:16686> to access the Jaeger UI.

Notice the exported metrics at <http://localhost:2222/>. Navigate to <http://localhost:9090> to access the Prometheus's expression browser.
