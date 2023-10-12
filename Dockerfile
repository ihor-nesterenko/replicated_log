FROM golang:1.21 AS builder

WORKDIR  /log

COPY . ./

ENV GOPROXY=direct GOSUMDB=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o replicated_log .

FROM alpine:latest AS release
WORKDIR /
COPY --from=builder /log/replicated_log /replicated_log

CMD ["/replicated_log"]
