FROM golang:1.24.4-alpine3.22 AS build
COPY go.mod go.sum /build/
WORKDIR /build
RUN go mod download
COPY cmd /build/cmd
COPY internal /build/internal
RUN go build -o knowledgebot -ldflags "-s -w -extldflags \"-static\"" ./cmd/knowledgebot

FROM alpine:3.22
COPY --from=build /build/knowledgebot /knowledgebot
COPY ui /var/lib/knowledgebot/ui
ENTRYPOINT ["/knowledgebot"]
CMD ["serve"]
