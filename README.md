# batch-rest-controller

Simple Batch applications controller by REST API.

## Getting started

requires MySQL database.  
edit config/database.json

```shell
$ go run create_tables.go
$ go run cli_add_api_client.go your_api_client_name
$ go run server.go
```

## API

### run job

URL: http://localhost:8000/jobs/run

## License

MIT
