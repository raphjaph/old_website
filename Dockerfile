FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /website
COPY . .

RUN go mod download
RUN go build

#FROM scratch
#
#WORKDIR /website
#COPY --from=builder /website .

ENTRYPOINT [ "./website"]
CMD ["run"]

