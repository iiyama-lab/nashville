package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/DaikiYamakawa/deepl-go"
	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape(client *deepl.Client) {
	// Request the HTML page.
	res, err := http.Get("https://openaccess.thecvf.com/content/CVPR2021/html/Wu_Greedy_Hierarchical_Variational_Autoencoders_for_Large-Scale_Video_Prediction_CVPR_2021_paper.html")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	translateResponse, err := client.TranslateSentence(context.Background(), doc.Find("#abstract").Text(), "EN", "JA")
	if err != nil {
		fmt.Printf("Failed to translate text:\n   %+v\n", err)
	} else {
		fmt.Printf("%+v\n", translateResponse)
	}

	// Find the review items
	doc.Find("#abstract").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Find("a").Text()
		fmt.Printf("Review %d: %s\n", i, title)
	})
}

func main() {

	client, err := deepl.New("https://api-free.deepl.com", nil)
	if err != nil {
		fmt.Printf("Failed to create client:\n   %+v\n", err)
	}

	ExampleScrape(client)
}
