package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"investing-scrapers/internal/scraper"
)

func main() {
	command := flag.String("cmd", "currencies", "Command")
	query := flag.String("q", "", "Search query")
	name := flag.String("name", "", "Instrument name/slug for detail")
	flag.Parse()

	switch *command {
	case "currencies":
		s := scraper.NewCurrencyScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create currency scraper")
			os.Exit(1)
		}
		rates, err := s.FetchAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(rates, "", "  ")
		fmt.Println(string(data))

	case "commodities":
		s := scraper.NewCommodityScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create commodity scraper")
			os.Exit(1)
		}
		quotes, err := s.FetchAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Note: %v\n", err)
			fmt.Println("[]")
			return
		}
		data, _ := json.MarshalIndent(quotes, "", "  ")
		fmt.Println(string(data))

	case "commodity-quote":
		if *name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required for commodity-quote (e.g. gold)")
			os.Exit(1)
		}
		s := scraper.NewCommodityQuoteScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create scraper")
			os.Exit(1)
		}
		detail, err := s.FetchByName(*name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(detail, "", "  ")
		fmt.Println(string(data))

	case "indices":
		s := scraper.NewIndexScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create index scraper")
			os.Exit(1)
		}
		quotes, err := s.FetchAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Note: %v\n", err)
			fmt.Println("[]")
			return
		}
		data, _ := json.MarshalIndent(quotes, "", "  ")
		fmt.Println(string(data))

	case "search":
		if *query == "" {
			fmt.Fprintln(os.Stderr, "Error: -q flag required for search")
			os.Exit(1)
		}
		s := scraper.NewSearchScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create search scraper")
			os.Exit(1)
		}
		resp, err := s.Search(*query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(data))

	case "index-quote":
		if *name == "" {
			fmt.Fprintln(os.Stderr, "Error: -name flag required for index-quote (e.g. us-30)")
			os.Exit(1)
		}
		s := scraper.NewIndexQuoteScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create scraper")
			os.Exit(1)
		}
		detail, err := s.FetchByName(*name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(detail, "", "  ")
		fmt.Println(string(data))

	case "funds":
		s := scraper.NewFundsScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create funds scraper")
			os.Exit(1)
		}
		quotes, err := s.FetchAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(quotes, "", "  ")
		fmt.Println(string(data))

	case "crypto":
		s := scraper.NewCryptoScraper()
		if s == nil {
			fmt.Fprintln(os.Stderr, "Failed to create crypto scraper")
			os.Exit(1)
		}
		quotes, err := s.FetchAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(quotes, "", "  ")
		fmt.Println(string(data))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", *command)
		fmt.Fprintf(os.Stderr, "Usage: %s -cmd=[currencies|commodities|commodity-quote|indices|search|funds|crypto] [-q=query] [-name=slug]\n", os.Args[0])
		os.Exit(1)
	}
}
