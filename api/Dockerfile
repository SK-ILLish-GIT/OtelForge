FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ internal/
COPY scripts/ scripts/
COPY api/ api/
RUN CGO_ENABLED=0 go build -o /server ./api/cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
