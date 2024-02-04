package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tmshv/ghstars/github"
	"github.com/tmshv/ghstars/set"

	tea "github.com/charmbracelet/bubbletea"
)

type GhstarsStartMsg struct{}
type GhstarsStopMsg struct{}

type AddStarMsg struct {
	star *github.GhStarV3
}

func GhStartFetch() tea.Msg {
	return GhstarsStartMsg{}
}

func GhStopFetch() tea.Msg {
	return GhstarsStopMsg{}
}

func Ghfetch(p *tea.Program) {
	languages := map[string]int{}
	topics := set.New[string]()
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}
	token := os.Getenv("GITHUB_TOKEN")
	gh := github.New(token)
	var total int
	for res := range gh.GetStars("tmshv") {
		star, err := res.Unwrap()
		if err != nil {
			log.Fatalf("Got error: %s", err)
		}
		total++

		// i := item{title: star.Repo.HTMLURL, desc: star.Repo.Description}
		p.Send(AddStarMsg{star: star})
		// 	item{title: "Terrycloth", desc: "In other words, towel fabric"},
		continue

		for _, topic := range star.Repo.Topics {
			topics.Add(topic)
		}

		fmt.Printf("%s\n", star.Repo.Name)
		fmt.Printf("%s\n", star.Repo.HTMLURL)
		fmt.Printf("%s\n", star.Repo.Description)

		ts := strings.Join(star.Repo.Topics, ", ")
		fmt.Printf("%s\n", ts)

		starredAt := star.StarredAt.Format("20060102")
		fmt.Printf("Starred At %s\n", starredAt)

		monthsPassed := int(time.Since(star.Repo.UpdatedAt).Hours() / 24 / 30)
		if monthsPassed != 0 {
			updatedAt := star.Repo.UpdatedAt.Format("20060102")
			fmt.Printf("Months since last update: %d (%s)\n", monthsPassed, updatedAt)
		}

		if star.Repo.Language != "" {
			fmt.Printf("Language: %s\n", star.Repo.Language)
			languages[star.Repo.Language]++
		}

		fmt.Println("")
	}

	p.Send(GhStopFetch())
	return

	fmt.Printf("Total %d\n", total)

	fmt.Println("Topics:")
	for _, t := range topics.Items() {
		fmt.Println(t)
	}

	fmt.Println("Languages:")
	type count struct {
		val   string
		count int
	}
	langs := make([]count, 0, len(languages))
	for lang, cnt := range languages {
		langs = append(langs, count{val: lang, count: cnt})
	}
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].count > langs[j].count
	})
	for _, pair := range langs {
		fmt.Printf("%s: %d\n", pair.val, pair.count)
	}
}
