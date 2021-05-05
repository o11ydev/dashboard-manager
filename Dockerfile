FROM golang:1.16 AS build
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

FROM gcr.io/distroless/static-debian10
COPY --from=build /go/bin/dashboard-manager /
CMD ["/dashboard-manager"]
