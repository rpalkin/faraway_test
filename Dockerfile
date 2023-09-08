FROM golang:1.20 as build

WORKDIR /app

COPY main.go go.sum go.mod .

RUN go vet $(go list ./... | grep -v /vendor/)

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-extldflags '-static'" -o app ./

FROM gcr.io/distroless/static-debian11

COPY --from=build /app/app /app

CMD ["/app"]
