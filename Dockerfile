FROM golang:1.17.1

WORKDIR /xschedule/builddir

COPY . .

RUN go mod download
RUN go build -o /main jaapie/xscheduleapi

WORKDIR /

ENV address=':8080'

ENV GIN_MODE=release


EXPOSE 8080

CMD ["./main"]