# Sample macaroon-based identity service

A sample identity service using
[macaroons](https://github.com/go-macaroon/macaroon) using the "form"
authentication mechanism with user/password based on a CSV file.

It provides two binaries:
- `macaroon-identity`: server providing form-based authentication with user/password
- `test-macaroon-identity`: a sample executable which features a client
  connecting to a service which requires authentication through a third-party
  authentication service (which is effectively `macaroon-identity`).

## Installing

With a `GOPATH` set, run

```bash
go get -v github.com/albertodonato/macaroon-identity/...
go install github.com/albertodonato/macaroon-identity/...
```

## Running

The server needs a CSV file containing credentials in the form `username,password`.

```bash
echo 'foo,bar' > credentials.csv
$GOPATH/bin/macaroon-identity
```

By default the server endpoint is <http://localhost:8081>.
