FROM golang:1.25-alpine AS build
WORKDIR /src
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY migrations/ migrations/
COPY queries/ queries/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/server /bin/server
COPY --from=build /go/bin/goose /bin/goose
COPY --from=build /src/migrations /migrations
COPY --from=build /etc/ssl/certs /etc/ssl/certs
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
