package main

import (
	"context"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const BotToken = "..."
const AuthorizedUser = "..."
const ApiEndpoint = ""
const ConnectionAttempts = 5
const RetryTime = 5 * time.Second

func initApi() (*tgbotapi.BotAPI, error) {
	i := ConnectionAttempts
	for {
		bot, err := tgbotapi.NewBotAPI(BotToken)
		i--
		if err == nil {
			return bot, nil
		} else {
			log.Printf("Connection failed with error %s", err.Error())
			if i == 0 {
				log.Printf("Attempts exhausted, returning")
				return bot, err
			} else {
				log.Printf("Retrying after timeout")
				time.Sleep(RetryTime)
			}
		}
	}
}

func main() {
	bot, err := initApi()
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	client := NewClient()
	ctx := context.Background()
	for update := range updates {
		if update.Message != nil {
			if update.Message.From.UserName != AuthorizedUser {
				continue
			}
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "clean":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please select zone to clean")
					zones, err := client.GetZoneCleaningCapabilityPresets(ctx)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					}
					buttons := make([]tgbotapi.InlineKeyboardButton, len(*zones)+1)
					allData := "all"
					buttons[0] = tgbotapi.InlineKeyboardButton{
						Text:         "All",
						CallbackData: &allData,
					}
					i := 1
					for _, value := range *zones {
						callbackData := value.ID
						buttons[i] = tgbotapi.InlineKeyboardButton{
							Text:         value.Name,
							CallbackData: &callbackData,
						}
						i++
					}
					msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
							buttons,
						},
					}
					bot.Send(msg)
				case "pause":
					err := client.PutBasicControlCapability(ctx, "pause")
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					}
				case "home":
					err := client.PutBasicControlCapability(ctx, "home")
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					}
				case "start":
					// Do nothing
				case "status":
					attrs, err := client.GetStateAttributes(ctx)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					}
					var status string
					var batteryLevel int
					for _, attr := range *attrs {
						switch attr.Class {
						case "StatusStateAttribute":
							status = attr.Value.(string)
						case "BatteryStateAttribute":
							batteryLevel = attr.Level
						}
					}
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Status: "+status+"\nBattery: "+strconv.Itoa(batteryLevel)+"%"))
				default:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command"))
				}
			}
		}
		if update.CallbackQuery != nil {
			if update.CallbackQuery.From.UserName != AuthorizedUser {
				continue
			}
			switch update.CallbackQuery.Data {
			case "all":
				err := client.PutBasicControlCapability(ctx, "start")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error()))
				}
			default:
				// TODO: input validation for UUID
				err := client.PutZoneCleaningCapabilityPresets(ctx, update.CallbackQuery.Data)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error()))
				}
			}
		}

	}
}
