FROM golang:1.18

WORKDIR /xschedule/builddir

COPY . .

RUN go mod download
RUN go build -o /main jaapie/xscheduleapi

WORKDIR /

COPY ./teachersList.json .

ENV address=':8080'

ENV GIN_MODE=release


EXPOSE 8080

CMD ["./main"]