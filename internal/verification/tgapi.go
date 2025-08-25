package verification

import (
	"context"
	"log"
	"os"

	"github.com/go-telegram/bot"
)

func CheckTGApi(string) bool {

	b, err := bot.New(os.Getenv("TG_TOKEN"))
	if err != nil {
		log.Fatalf("create bot error: %v", err)
	}

	b.Start(context.TODO())

	return true
}
