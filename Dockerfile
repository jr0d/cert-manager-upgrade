FROM golang:1.15 as build

WORKDIR /usr/local/src/cert-manager-upgrade

COPY . .

RUN go build -o /cert-manager-upgrade .

FROM ubuntu:18.04

COPY --from=build /cert-manager-upgrade /cert-manager-upgrade

ENTRYPOINT ["/cert-manager-upgrade"]
