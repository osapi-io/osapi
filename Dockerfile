FROM golang:1.25 AS builder

ENV CGO_ENABLED=0

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o osapi .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /build/osapi /usr/local/bin/osapi

ENTRYPOINT ["osapi"]
