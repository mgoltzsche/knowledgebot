FROM golang:1.24.4-alpine3.22 AS go

FROM go AS linter
RUN apk add --update --no-cache git
ARG GOLANGCILINT_VERSION=v2.1.6
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$GOLANGCILINT_VERSION

FROM go AS builddeps
COPY go.mod go.sum /build/
WORKDIR /build
RUN go mod download
COPY cmd /build/cmd
COPY internal /build/internal

FROM builddeps AS test
RUN go test ./...

FROM builddeps AS lint
COPY --from=linter /go/bin/golangci-lint /go/bin/
RUN golangci-lint run ./...

FROM builddeps AS build
RUN go build -o knowledgebot -ldflags "-s -w -extldflags \"-static\"" ./cmd/knowledgebot

FROM alpine:3.22
COPY --from=build /build/knowledgebot /knowledgebot
COPY ui /var/lib/knowledgebot/ui
ENTRYPOINT ["/knowledgebot"]
CMD ["serve"]
