# Observability via OpenTelemtry Go - Quick Start

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
