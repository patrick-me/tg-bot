FROM golang:1.20-alpine

WORKDIR /bot
#ENV TELEGRAM_APITOKEN=!BotFather-tkn!
#ENV SERVER_ADDR=localhost:50051

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY bot.go ./
COPY proxyClient.go ./
COPY proto ./proto

RUN go build -o /tg-bot

CMD [ "/tg-bot" ]
