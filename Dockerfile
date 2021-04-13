################
FROM golang:1.16 as builder
RUN go version

WORKDIR /go/src/github.com/noelruault/auction-bid-tracker/
# The first dot refers to the local path itself, the second one points the WORKDIR path
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build ./cmd/auction-api/

################
FROM alpine

COPY --from=builder /go/src/github.com/noelruault/auction-bid-tracker/auction-api/ auction-api

RUN chmod +x ./auction-api

ENTRYPOINT ENV=$ENV ./auction-api $ENV_ARGS
