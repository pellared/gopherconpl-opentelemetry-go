# Observability via OpenTelemtry Go - Quick Start

[Presentation](https://docs.google.com/presentation/d/1ir9eyBLxO4n3zPcOPhxJkZJ9cXJSo7KUhuRAadG7z8Q/edit?usp=sharing)

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
