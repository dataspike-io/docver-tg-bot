package main

import (
	"context"
	"fmt"
	"github.com/dataspike-io/docver-sdk-go"
	"github.com/dataspike-io/docver-tg-bot/internal/cache"
	"github.com/dataspike-io/docver-tg-bot/internal/handlers"
	"github.com/dataspike-io/docver-tg-bot/pkg/telegram_bot"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	//"github.com/dataspike-io/docver-tg-bot/pkg/gateways"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	cfg := newConfig()
	ctx, cancel := context.WithCancel(context.Background())

	memoryCache, err := cache.NewMemoryCache(0)

	dataspikeClient := dataspike.NewDataspikeClient(dataspike.WithEndpoint(cfg.dataspikeUrl), dataspike.WithToken(cfg.DataspikeToken.RawString()))
	err = createWebhook(cfg.webhookUrl, dataspikeClient)
	if err != nil {
		log.Fatalf("failed to create webhook: %s", err)
	}

	bot, err := telegram_bot.NewTelegramBot(cfg.TelegramToken.RawString(), dataspikeClient, memoryCache)
	if err != nil {
		log.Fatalf("failed to create telegram bot: %s", err)
	}
	go bot.Start(ctx, cfg.telegramOffset, cfg.telegramTimeout)

	handler := handlers.NewTgBotHandler(bot)
	mux := http.NewServeMux()
	mux.Handle(cfg.webhookPath, handler)

	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.httpPort), mux)
	if err != nil {
		log.Fatalf("failed to listen server: %s", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	cancel()
}

func createWebhook(webhookUrl string, client dataspike.IDataspikeClient) error {
	webhookData, err := client.ListWebhooks()
	if err != nil {
		return err
	}

	for _, w := range webhookData.Webhooks {
		if w.WebhookUrl == webhookUrl {
			return nil
		}
	}

	return client.CreateWebhook(&dataspike.WebhookCreate{WebhookUrl: webhookUrl, EventTypes: []string{"DOCVER", "DOCVER_CHECKS"}, Enabled: true})
}
