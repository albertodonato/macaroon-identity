# Sample macaroon-based identity service

[![GoDoc](https://godoc.org/github.com/albertodonato/macaroon-identity?status.svg)](https://godoc.org/github.com/albertodonato/macaroon-identity)
[![Build status](https://img.shields.io/travis/albertodonato/macaroon-identity.svg)](https://travis-ci.org/albertodonato/macaroon-identity)

A sample identity service based on
[macaroon](https://github.com/go-macaroon/macaroon) using the "form"
authentication mechanism with user/password based on a CSV file.

It provides two binaries:
- `macaroon-identity`: server providing form-based authentication with user/password
- `test-macaroon-identity`: a sample executable which features a client
  connecting to a service which requires authentication through a third-party
  authentication service (which is effectively `macaroon-identity`).

*NOTE*: the `macaroon-identity` server is meant as a sample service and not
intended for any production use.


## Installing

With a `GOPATH` set, run

```bash
go get -v github.com/albertodonato/macaroon-identity/...
```

## Running

The server needs a CSV file containing credentials in the form
`username,password,groups`.  The `groups` column is a space-separated list of
groups that the user is part of and it's optional.

```bash
echo 'user1,pass2,grp1 grp2' > credentials.csv
$GOPATH/bin/macaroon-identity
```

By default the server endpoint is <http://localhost:8081>.


## Testing a request

The `test-macaroon-identity` command starts the identity service and a test
service which responds to `GET /` requests , requiring authentication through
the identity service.

If `-username` and `-password` options are passed, the command also issues a
request to the target service, to test the authentication process.
