FROM golang:1.21-bullseye as builder

WORKDIR /build

COPY . .
RUN go mod download
RUN go build -o trader ./cmd

FROM debian:bullseye as runner

RUN apt update && apt install -y ca-certificates

WORKDIR /runner
COPY --from=builder /build/trader /runner

ENTRYPOINT [ "/runner/trader" ]
