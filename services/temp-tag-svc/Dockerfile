FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/temp-tag-svc ./cmd/temp-tag-svc

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/temp-tag-svc /temp-tag-svc
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/temp-tag-svc"]
