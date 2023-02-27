## tg-bot
Bot connects to tg-bot and provide messages to proxyServer to process them

to local run type: `go run server/proxyServer.go`


## proxy server
Server will serve some logic by user message input

to local run type: `go run bot.go`


## docker compose

Combines 2 docker files and connect them together
- [Dockerfile](Dockerfile)
- [Dockerfile-server](Dockerfile-server)

Also you could provide required ENV variables

      SERVER_ADDR: "proxy_server:50051"
      TELEGRAM_APITOKEN: "!BotFather-tkn!"

to local run type: `docker-compose up -d`