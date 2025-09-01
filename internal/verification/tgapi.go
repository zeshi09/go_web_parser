package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func main() {
	// Creds: в проде — os.Getenv, не хардкодь. Для sec: rotate app_id/hash periodically.
	appID := 123456                 // Замени
	appHash := "your_api_hash_here" // Замени
	phone := "+1234567890"          // Burner для user auth
	// botToken := "your_bot_token_here" // Для bot auth, но не для createChannel

	// Клиент. SessionStorage: encrypt с kms или age для защиты от leaks.
	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: "tg_session.json"},
		// Proxy для evasion: telegram.ProxySOCKS5("socks5://127.0.0.1:9050", "")
	})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Status check: если уже auth'ed, skip flow.
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("status check failed: %w", err)
		}
		if status.Authorized {
			fmt.Println("Сессия валидна — телефон не нужен!")
		} else {
			// User auth: нужен phone. В red team: automate с SMS API.
			codeAuthenticator := auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
				fmt.Printf("Код отправлен через: %T\n", sentCode.Type)
				fmt.Print("Введите код: ")
				var code string
				_, err := fmt.Scanln(&code)
				return code, err
			})
			flow := auth.NewFlow(auth.Constant(phone, "", codeAuthenticator), auth.SendCodeOptions{})

			// Альтернатива: bot auth (uncomment, но не для create).
			// flow = auth.NewFlow(auth.Bot(botToken), auth.SendCodeOptions{})

			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("auth failed: %w (check proxy/virtual num)", err)
			}
		}

		api := client.API()

		// Create channel: только для user.
		req := &tg.ChannelsCreateChannelRequest{
			Title:     "MyRedTeamChannelCheckCheckCheck",
			About:     "C2 test. OPSEC: obfuscate metadata!",
			Broadcast: true,
			Megagroup: false,
		}

		updates, err := api.ChannelsCreateChannel(ctx, req)
		if err != nil {
			return fmt.Errorf("create failed: %w (возможно, bot auth или ban)", err)
		}

		fmt.Printf("Канал готов: %+v\n", updates)

		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // Не логи creds!
		os.Exit(1)
	}
}
