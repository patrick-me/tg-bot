FROM golang:1.20-alpine

WORKDIR /bot
#ENV TELEGRAM_APITOKEN=tkn

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /tg-bot

CMD [ "/tg-bot" ]