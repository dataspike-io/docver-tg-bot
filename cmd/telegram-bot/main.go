package main

import (
	"context"
	"fmt"
	"github.com/dataspike-io/docver-sdk-go"
	"github.com/dataspike-io/docver-tg-bot/internal/cache"
	"github.com/dataspike-io/docver-tg-bot/internal/handlers"
	"github.com/dataspike-io/docver-tg-bot/pkg/telegram_bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken.RawString())
	if err != nil {
		log.Fatalf("failed to create BotAPI: %s", err)
	}

	dsBot, err := telegram_bot.NewTelegramBot(bot, dataspikeClient, memoryCache)
	if err != nil {
		log.Fatalf("failed to create telegram dsBot: %s", err)
	}
	go dsBot.Start(ctx, cfg.telegramOffset, cfg.telegramTimeout)

	handler := handlers.NewTgBotHandler(dsBot)
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
