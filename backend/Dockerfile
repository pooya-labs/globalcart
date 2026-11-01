FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/globalcart-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog-import ./cmd/catalog-import

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/globalcart-api /globalcart-api
COPY --from=build /out/catalog-import /catalog-import
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/globalcart-api"]
