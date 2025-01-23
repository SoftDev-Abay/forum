FROM golang:1.22-alpine3.20 as base

RUN apk add build-base 
WORKDIR /web
COPY . .
RUN go build -o forum ./cmd/web/

FROM alpine:3.20
WORKDIR /web
COPY --from=base /web/ /web/

CMD ["./forum"]