package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
)

func isYoutubeLink(rawurl string) bool {
	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}
	return u.Host == "www.youtube.com" || u.Host == "youtube.com" || u.Host == "youtu.be"
}

func downloadYoutubeVideo(videoID string) (string, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoID)
	if err != nil {
		return "", err
	}

	formats := video.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return "", err
	}

	tempFile, err := ioutil.TempFile("", "video-*.mp4")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, stream)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func main() {
	// Loading environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Bot token is required.")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if isYoutubeLink(update.Message.Text) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Processing...")
			processingMessage, _ := bot.Send(msg)

			videoID := update.Message.Text
			filePath, err := downloadYoutubeVideo(videoID)
			if err != nil {
				log.Println("Error downloading video:", err)
				continue
			}
			defer os.Remove(filePath)

			audio := tgbotapi.NewAudioUpload(update.Message.Chat.ID, filePath)
			log.Println("Sending audio...")
			log.Println("audio path:", filePath)
			bot.Send(audio)

			bot.DeleteMessage(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, processingMessage.MessageID))

		} else if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, I'm a Youtube downloader bot. Send me a link to a Youtube video and I'll send you the audio")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "The link that you sent is not a Youtube link. Send me a link to a Youtube video and I'll send you the audio")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}
