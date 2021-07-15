package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
)

type Info struct {
	Name         string          `json:"name"`
	Rank         string          `json:"rank"`
	RegisteredAt string          `json:"registered_at"`
	LastLogin    string          `json:"last_login"`
	Coins        string          `json:"coins"`
	Friends      string          `json:"friends"`
	GameStats    map[string]Game `json:"stats"`
}

type Game struct {
	Image  string `json:"image"`
	Wins   string `json:"wins"`
	Points string `json:"points"`
}

var tableInfo []string
var Games map[string]Game

func check(err error) {
	if err != nil {
		log.Fatalf("Exception: %v", err)
	}
}

func main() {
	games := make(map[string]Game)

	app := fiber.New(fiber.Config{
		Prefork:      true,
		ServerHeader: "CraftRiseAPI",
	})

	v1 := app.Group("/api/v1")
	v1.Get("/user/:username", func(c *fiber.Ctx) error {
		resp, err := http.Get("https://www.craftrise.com.tr/oyuncu/" + c.Params("username"))
		check(err)

		if resp.Request.URL.String() == "https://www.craftrise.com.tr/" {
			return c.Status(404).JSON(&fiber.Map{
				"success": 0,
				"message": "Player not found!",
			})
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		check(err)

		doc.Find("table > tbody > tr > td[style=\"text-align: right;\"]").Each(func(index int, item *goquery.Selection) {
			tableInfo = append(tableInfo, item.Text())
		})

		doc.Find(".gameStats").Each(func(index int, item *goquery.Selection) {
			var gameInfo []string
			item.Find("li > span").Each(func(index int, item *goquery.Selection) {
				gameInfo = append(gameInfo, item.Text())
			})

			image, _ := item.Find("img").Attr("src")
			games[item.Find("h3").Text()] = Game{
				Image:  image,
				Wins:   gameInfo[0],
				Points: gameInfo[1],
			}
		})

		return c.JSON(Info{
			Name:         doc.Find("p[style=\"margin:0; padding: 0; font-size: 300%;\"]").Text(),
			Rank:         doc.Find(".rankButton[style=\"background-color: #4C4C4C\"] > p").Text(),
			RegisteredAt: strings.TrimSpace(tableInfo[0]),
			LastLogin:    strings.TrimSpace(tableInfo[1]),
			Coins:        tableInfo[2],
			Friends:      tableInfo[3],
			GameStats:    games,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app.Listen(":" + port)
}
