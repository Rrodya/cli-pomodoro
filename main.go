package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/gen2brain/beeep"
)

const (
	emptySimbolProgress  = "░"
	filledSimbolProgress = "█"
	fullProgressWidth    = 40
)

const (
	workEmoji  = "🍅"
	breakEmoji = "☕"
	pauseEmoji = "⏸️"
)

var colors = map[string]string{
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"red":     "\033[31m",
	"blue":    "\033[34m",
	"default": "\033[0m",
}

var progressionLevelPercentCheckpoint = map[string]int{
	"low":    50,
	"medium": 80,
	"high":   100,
}

type TimeConfig struct {
	Duration      int
	Mode          string
	Session       int
	TotalSessions int
}

func main() {
	workTime := flag.Int("work", 25, "Work part time")
	breakTime := flag.Int("break", 5, "Brear part time")
	sessionCount := flag.Int("session", 4, "Session count")

	flag.Parse()

	fmt.Printf("Кол-во сессий: %d\n\n", *sessionCount)

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	command := make(chan string)

	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				return
			}

			if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC {
				command <- "exit"
			} else if char == 'p' || char == 'P' {
				command <- "pause"
			}
		}
	}()

	for session := 1; session <= *sessionCount; session++ {
		config := TimeConfig{
			Session:       session,
			TotalSessions: *sessionCount,
			Duration:      *workTime,
			Mode:          "work",
		}

		if !runTimer(config, command) {
			fmt.Printf("\n\nТаймер остановлен! До скорого!")
			break
		}

		config.Mode = "break"
		config.Duration = *breakTime

		if session < *sessionCount {
			beeep.Alert("CLI Pomodoro", "Работа закончилась! Время отдыха!", "")
			fmt.Printf("\n\nРабота закончилась! Время отдыха\n\n")
		} else {
			beeep.Alert("CLI Pomodoro", "🎉 Все сессии завершены!", "")
			fmt.Printf("\n\n🎉 Все сессии завершены!\n")
		}

		if session < *sessionCount {
			if !runTimer(config, command) {
				fmt.Printf("\n\nТаймер остановлен! До скорого!")
				break
			}
			beeep.Alert("CLI Pomodoro", "Отдых закончился! Время работы", "")
			fmt.Printf("\n\nОтдых закончился! Время работы\n\n")
		}
	}

}

func runTimer(config TimeConfig, commandChan <-chan string) bool {
	fullTime := config.Duration
	countDownSec := fullTime
	isPaused := false
	ticker := time.NewTicker(1 * time.Second)

	var emoji string
	if config.Mode == "work" {
		emoji = workEmoji
	} else {
		emoji = breakEmoji
	}

	for {
		select {
		case cmd := <-commandChan:
			switch cmd {
			case "exit":
				return false
			case "pause":
				isPaused = !isPaused
			}
		case <-ticker.C:
			pauseIndicator := " "
			if isPaused {
				pauseIndicator = pauseEmoji + " "
			}
			var progress float32 = (float32(fullTime) - float32(countDownSec)) / float32(fullTime)
			filledProgressPart := int(progress * float32(fullProgressWidth))
			emptyProgressPart := fullProgressWidth - filledProgressPart

			min := countDownSec / 60
			sec := countDownSec % 60

			currentColor := colors[""]

			progressPercent := int(progress * 100)

			if progressPercent < progressionLevelPercentCheckpoint["low"] {
				currentColor = colors["green"]
			} else if progressPercent < progressionLevelPercentCheckpoint["medium"] {
				currentColor = colors["yellow"]
			} else if progressPercent < progressionLevelPercentCheckpoint["high"] {
				currentColor = colors["red"]
			} else {
				currentColor = colors["default"]
			}

			if config.Mode == "break" {
				currentColor = colors["blue"]
			}

			var currentStatusSession string
			if config.Mode == "work" {
				currentStatusSession = fmt.Sprintf(" Сессия %d/%d ", config.Session, config.TotalSessions)
			} else {
				currentStatusSession = " Перерыв "
			}

			fmt.Printf(
				"\033[2K\r%s%s ["+"%s"+
					strings.Repeat(
						filledSimbolProgress, int(filledProgressPart))+
					strings.Repeat(emptySimbolProgress, emptyProgressPart)+"%s"+
					"] %02d:%02d %s",
				emoji, currentStatusSession, currentColor, colors["default"], min, sec, pauseIndicator,
			)

			if !isPaused {
				countDownSec--
			}

			if countDownSec == -1 {
				ticker.Stop()
				return true
			}
		}
	}
}
