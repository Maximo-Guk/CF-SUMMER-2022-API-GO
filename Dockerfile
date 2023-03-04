## Build
FROM golang:1.20-bullseye AS build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY private.pem public.pem README.txt ./
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/app ./...

## Deploy
FROM debian:bullseye-slim

WORKDIR /
COPY --from=build /usr/src/app/private.pem .
COPY --from=build /usr/src/app/public.pem .
COPY --from=build /usr/src/app/README.txt .
COPY --from=build /usr/local/bin/app /usr/local/bin/app
EXPOSE 80

ENTRYPOINT ["app"]
