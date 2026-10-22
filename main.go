package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pingbot/internal"
	"time"

	"github.com/go-telegram/bot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbPath := os.Getenv("DB_PATH")
	conn, err := internal.InitDB(dbPath, 1*time.Second)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	ctx = context.WithValue(ctx, "conn", conn)

	opts := []bot.Option{
		bot.WithMessageTextHandler("reg", bot.MatchTypeCommand, internal.RegisterUserHandler),
		bot.WithMessageTextHandler("add", bot.MatchTypeCommand, internal.RegisterDomainHandler),
		bot.WithMessageTextHandler("my", bot.MatchTypeCommand, internal.MyDomainsHandler),
		bot.WithMessageTextHandler("del", bot.MatchTypeCommand, internal.DeleteDomainHandler),
	}

	token := os.Getenv("TOKEN")
	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatalf("create bot: %v", err)
	}

	_, err = b.GetMe(ctx)
	if err != nil {
		log.Fatalf("getMe: %v", err)
	}

	b.Start(ctx)
}
