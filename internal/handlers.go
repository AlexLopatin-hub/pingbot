package internal

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type User struct {
	ID            int64
	Username      string
	ActiveDomains []string
}

func RegisterUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	conn, ok := ctx.Value("conn").(Conn)
	if !ok {
		log.Panic("get connection from context")
	}

	message := update.Message

	user := User{
		ID:            message.From.ID,
		Username:      message.From.Username,
		ActiveDomains: make([]string, 0),
	}

	err := conn.AddUser(user)

	if err != nil {
		log.Printf("fialed to add user to db: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   InternalErrorText,
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   fmt.Sprintf("Успешно зарегистрирован пользователь %s, id=%d", user.Username, user.ID),
	})
}

func RegisterDomainHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	conn, ok := ctx.Value("conn").(Conn)
	if !ok {
		log.Panic("get connection from context")
	}

	message := update.Message

	domain := strings.TrimPrefix(message.Text, "/add ")
	matches, err := regexp.MatchString(`^https://([a-z0-9-]+\.)+[a-z]{2,}$`, domain)
	if err != nil {
		log.Panicf("compile regexp: %v", err)
	}

	if !matches {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   fmt.Sprintf("Некорректный домен: %s", domain),
		})
		return
	}

	userID := message.From.ID

	err = conn.AddDomain(userID, domain)
	if err != nil {
		switch err {
		case ErrNotFound:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   UserNotFoundText,
			})
			return
		default:
			log.Printf("failed to add domain to db: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   InternalErrorText,
			})
			return
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   fmt.Sprintf("Домен %s добавлен к отслеживаемым", domain),
	})
}

func MyDomainsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	conn, ok := ctx.Value("conn").(Conn)
	if !ok {
		log.Panic("get connection from context")
	}

	message := update.Message

	user, err := conn.GetUser(message.From.ID)
	if err != nil {
		switch err {
		case ErrNotFound:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   UserNotFoundText,
			})
			return
		default:
			log.Printf("failed to get user from db: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   InternalErrorText,
			})
			return
		}
	}

	domains := user.ActiveDomains
	if len(domains) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "В данный момент ни один домен не отслеживается",
		})
		return
	}

	sb := strings.Builder{}
	sb.WriteString("Отслеживаемые домены:\n")
	for i, v := range domains {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, v))
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   sb.String(),
	})
}

func DeleteDomainHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	conn, ok := ctx.Value("conn").(Conn)
	if !ok {
		log.Panic("get connection from context")
	}

	message := update.Message

	domain := strings.TrimPrefix(message.Text, "/del ")
	matches, err := regexp.MatchString(`^https://([a-z0-9-]+\.)+[a-z]{2,}$`, domain)
	if err != nil {
		log.Panicf("compile regexp: %v", err)
	}

	if !matches {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   fmt.Sprintf("Некорректный домен: %s", domain),
		})
		return
	}

	err = conn.DeleteDomain(message.From.ID, domain)
	if err != nil {
		switch err {
		case ErrNotFound:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   DomainNotFoundText,
			})
			return
		default:
			log.Printf("failed to delete domain from db: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: message.Chat.ID,
				Text:   InternalErrorText,
			})
			return
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   "Домен удалён из отслеживаемых",
	})
}
