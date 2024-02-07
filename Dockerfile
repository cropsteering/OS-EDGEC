FROM golang:1.21-alpine3.18 as builder
WORKDIR /app
ADD . /app
ADD ./config-docker ./config.go
RUN go mod download
RUN go build -o controller .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/controller /app/
CMD ["/app/controller"]

