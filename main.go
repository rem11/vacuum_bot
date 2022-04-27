package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ConnectionAttempts = 5
const RetryTime = 5 * time.Second

var configFile = flag.String("config", "/etc/vacuum_bot.json", "Config file path")

type Config struct {
	ApiUrl          string   `json:"apiUrl"`
	BotToken        string   `json:"botToken"`
	AuthorizedUsers []string `json:"authorizedUsers"`
}

func initApi(config *Config) (*tgbotapi.BotAPI, error) {
	i := ConnectionAttempts
	for {
		bot, err := tgbotapi.NewBotAPI(config.BotToken)
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

func validateConfig(config Config) error {
	if config.ApiUrl == "" {
		err := errors.New("apiUrl is missing in config")
		log.Print(err.Error())
		return err
	}
	if config.BotToken == "" {
		err := errors.New("botToken is missing in config")
		log.Print(err.Error())
		return err
	}
	if config.AuthorizedUsers == nil {
		err := errors.New("authorizedUsers are missing in config")
		log.Print(err.Error())
		return err
	}
	return nil
}

func readConfig() (*Config, error) {
	flag.Parse()
	file, err := os.Open(*configFile)
	defer file.Close()
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	err = validateConfig(config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func setCommands(bot *tgbotapi.BotAPI) error {
	_, err := bot.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "zones",
			Description: "List zones for cleanup",
		},
		tgbotapi.BotCommand{
			Command:     "pause",
			Description: "Pause",
		},
		tgbotapi.BotCommand{
			Command:     "home",
			Description: "Go back to the dock",
		},
		tgbotapi.BotCommand{
			Command:     "status",
			Description: "Display status",
		},
	))
	if err != nil {
		log.Print(err.Error())
		return err
	}
	return nil
}

func isAuthorizedUser(userName string, config *Config) bool {
	for _, authorizedUser := range config.AuthorizedUsers {
		if userName == authorizedUser {
			return true
		}
	}
	return false
}

func processUpdates(bot *tgbotapi.BotAPI, config *Config) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	client := NewClient(config.ApiUrl)
	ctx := context.Background()
	for update := range updates {
		if update.Message != nil {
			if !isAuthorizedUser(update.Message.From.UserName, config) {
				continue
			}
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "zones":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available zones:")
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
						callbackData := value.ID + "|" + value.Name
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
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Pausing"))
					}
				case "home":
					err := client.PutBasicControlCapability(ctx, "home")
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Going back to the dock"))
					}
				case "start":
					// Do nothing
				case "status":
					attrs, err := client.GetStateAttributes(ctx)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					} else {
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
					}
				default:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command"))
				}
			}
		}
		if update.CallbackQuery != nil {
			if !isAuthorizedUser(update.CallbackQuery.From.UserName, config) {
				continue
			}
			callbackData := strings.Split(update.CallbackQuery.Data, "|")
			if callbackData[0] == "all" {
				err := client.PutBasicControlCapability(ctx, "start")
				if err != nil {
					bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, err.Error()))
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error()))
				} else {
					bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Starting cleanup"))
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Starting cleanup"))
				}
			} else {
				// TODO: input validation for UUID
				err := client.PutZoneCleaningCapabilityPresets(ctx, callbackData[0])
				if err != nil {
					bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, err.Error()))
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, err.Error()))
				} else {
					bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Starting cleanup for zone: "+callbackData[1]))
					bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Starting cleanup for zone: "+callbackData[1]))
				}
			}
		}

	}
}

func main() {
	config, err := readConfig()
	if err != nil {
		log.Panic(err)
	}
	bot, err := initApi(config)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	setCommands(bot)
	processUpdates(bot, config)
}
