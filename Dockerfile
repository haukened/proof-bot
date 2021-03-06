FROM golang:stretch AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v
RUN go build -o app .

FROM keybaseio/client:stable
WORKDIR /home/keybase/
RUN mkdir /home/keybase/proof/
COPY --from=builder /go/src/app/app .
ENV KEYBASE_SERVICE=1
CMD ["./app"]
