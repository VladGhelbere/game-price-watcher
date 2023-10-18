package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	colly "github.com/gocolly/colly"
)

const (
	AKS_DOMAIN            = "www.allkeyshop.com"
	STEAM_DOMAIN          = "store.steampowered.com"
	AKS_BASE_URL          = "https://www.allkeyshop.com"
	PRICE_LOOKUP_ENDPOINT = "/blog/catalogue/category-pc-games-all/search-"
	BEST_PRICE_QUERY      = "li.search-results-row:first-of-type div.search-results-row-price"
)

type (
	SteamGame struct {
		Id             int64  `json:"-"`
		Name           string `json:"name,omitempty"`
		ReviewScore    int    `json:"review_score,omitempty"`
		ReviewDesc     string `json:"review_desc,omitempty"`
		ReviewsTotal   string `json:"reviews_total,omitempty"`
		ReviewsPercent int    `json:"reviews_percent,omitempty"`
		ReleaseDate    any    `json:"release_date,omitempty"`
		ReleaseString  string `json:"release_string,omitempty"`
		Subs           []struct {
			Price string `json:"price,omitempty"`
		} `json:"subs,omitempty"`
		Type        string   `json:"type,omitempty"`
		Screenshots []string `json:"screenshots,omitempty"`
		ReviewCSS   string   `json:"review_css,omitempty"`
		Priority    int      `json:"priority,omitempty"`
		Added       int      `json:"added,omitempty"`
		Rank        int      `json:"rank,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		IsFreeGame  bool     `json:"is_free_game,omitempty"`
		Win         int      `json:"win,omitempty"`
		Mac         int      `json:"mac,omitempty"`
		Linux       int      `json:"linux,omitempty"`
		BestPrice   float64  `json:"-"`
	}
)

func getUserWishlist(userId string) ([]SteamGame, error) {
	c := colly.NewCollector(colly.AllowedDomains(STEAM_DOMAIN))

	c.OnError(func(r *colly.Response, e error) {
		fmt.Printf("Error while scraping: %s\n", e.Error())
	})

	// gameId -> gameInfo map
	var steamGamesMap map[string]SteamGame
	var steamGames []SteamGame

	c.OnResponse(func(r *colly.Response) {
		err := json.Unmarshal(r.Body, &steamGamesMap)
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	c.Visit(fmt.Sprintf("https://store.steampowered.com/wishlist/profiles/%s/wishlistdata", userId))

	// move map "id" data to slice
	for gameId, game := range steamGamesMap {
		err := errors.New("")
		game.Id, err = strconv.ParseInt(gameId, 0, 64)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		steamGames = append(steamGames, game)
	}

	return steamGames, nil
}

func lookupBestPrice(steamGames []SteamGame) error {
	c := colly.NewCollector(colly.AllowedDomains(AKS_DOMAIN))

	aksRequests := 0
	for i, gameData := range steamGames {
		bestPrice := ""
		gameNameSanitized := strings.Replace(gameData.Name, " ", "+", -1)

		c.OnError(func(r *colly.Response, e error) {
			fmt.Printf("Error while scraping: %s\n", e.Error())
			aksRequests++
		})

		c.OnHTML(BEST_PRICE_QUERY, func(h *colly.HTMLElement) {
			bestPrice = strings.TrimSuffix(strings.Replace(strings.Replace(h.Text, " ", "", -1), "\n", "", -1), "€")
		})

		c.Visit(fmt.Sprintf("%s%s%s", AKS_BASE_URL, PRICE_LOOKUP_ENDPOINT, gameNameSanitized))

		if bestPrice == "" {
			fmt.Println(errors.New(fmt.Sprintf("ERROR - No price found for game: %s", gameData.Name)))
			continue
		}
		bestPriceFloat, err := strconv.ParseFloat(bestPrice, 64)
		if err != nil {
			fmt.Println(errors.New(fmt.Sprintf("ERROR - Cannot convert price to float for game %s: %s", gameData.Name, err)))
			continue
		}
		steamGames[i].BestPrice = bestPriceFloat
		fmt.Println("Best price for: ", gameData.Name, "is: ", steamGames[i].BestPrice)
		aksRequests++
		if aksRequests >= 5 {
			time.Sleep(30 * time.Second)
			aksRequests = 0
		}
	}
	return nil
}

func main() {
	// define SteamGame map object & get user wishlist
	var wishlistedGames []SteamGame
	wishlistedGames, err := getUserWishlist("76561198062700091")
	if err != nil {
		fmt.Println(err)
		return
	}

	// iterate through wishlisted games & grab best prices for each
	err = lookupBestPrice(wishlistedGames)
	if err != nil {
		fmt.Println(err)
		return
	}

	sort.Slice(wishlistedGames, func(i, j int) bool {
		return wishlistedGames[i].BestPrice < wishlistedGames[j].BestPrice
	})

	fmt.Println("Price", "\t", "Game")

	for _, game := range wishlistedGames {
		fmt.Println(game.BestPrice, "\t", game.Name)
	}
}
