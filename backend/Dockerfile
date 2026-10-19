FROM golang:1.27.2-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server .

FROM gcr.io/distroless/base-debian12
ENV GIN_MODE=release
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
