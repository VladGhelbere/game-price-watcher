package main

import (
	"errors"
	"fmt"
	"strings"

	colly "github.com/gocolly/colly"
)

const (
	DOMAIN                = "www.allkeyshop.com"
	BASE_URL              = "https://www.allkeyshop.com"
	PRICE_LOOKUP_ENDPOINT = "/blog/catalogue/category-pc-games-all/search-"
	BEST_PRICE_QUERY      = "li.search-results-row:first-of-type div.search-results-row-price"
)

func lookupBestPrice(gameName string) (string, error) {
	bestPrice := ""
	gameNameSanitized := strings.Replace(gameName, " ", "+", -1)

	c := colly.NewCollector(colly.AllowedDomains(DOMAIN))

	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("Error while scraping: %s\n", e.Error())
	})

	c.OnHTML(BEST_PRICE_QUERY, func(h *colly.HTMLElement) {
		bestPrice = (strings.Replace(strings.Replace(h.Text, " ", "", -1), "\n", "", -1))
	})

	c.Visit(fmt.Sprintf("%s%s%s", BASE_URL, PRICE_LOOKUP_ENDPOINT, gameNameSanitized))

	if bestPrice == "" {
		return "", errors.New(fmt.Sprintf("ERROR - No price found for game: %s", gameName))
	}
	return bestPrice, nil
}

func main() {
	wishlistedGames := []string{
		"Superhot",
		"Assassin's Creed 4 Black Flag",
		"Assassin's Creed Mirage",
	}

	for _, gameName := range wishlistedGames {
		bestPrice, err := lookupBestPrice(gameName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Best price for: ", gameName, "is: ", bestPrice)
	}
}
