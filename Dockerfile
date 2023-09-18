FROM golang:alpine AS builder
LABEL builder=true
WORKDIR /build

ADD go.mod .

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o main main.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/main /build/main
COPY --from=builder /build/postgres/migrations/ /build/postgres/migrations/
EXPOSE 3000
CMD ["./main"]