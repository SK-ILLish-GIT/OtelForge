FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ internal/
COPY scripts/ scripts/
COPY worker/ worker/
RUN CGO_ENABLED=0 go build -o /worker ./worker/cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates openssh-client
COPY --from=build /worker /worker
ENTRYPOINT ["/worker"]
