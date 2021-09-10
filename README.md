# Observability via OpenTelemtry Go - Quick Start

Run the database:

```sh
docker run --name todo-db -p 5432:5432 -e POSTGRES_PASSWORD=pswd -v $(pwd)/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql -d --rm postgres:13-alpine
```

Run the service:

```sh
cd todoservice
go run .
```

Run the CLI app:

```sh
cd todo
go run . add "important work"
go run . list
go run .
```
