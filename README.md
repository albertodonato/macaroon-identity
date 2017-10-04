# Sample macaroon-based identity service

## Installation

Run

```bash
go get -v github.com/albertodonato/macaroon-identity/...
go install github.com/albertodonato/macaroon-identity/...
```

## Running

The server needs a CSV file containing credentials in the form `username,password`.

```bash
echo 'foo,bar' > credentials.csv
macaroon-identity
```

By default the server endpoint is <http://localhost:8081>.
