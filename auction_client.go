package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Bid struct {
	Bidder string `json:"bidder"`
	Amount int    `json:"amount"`
}

type OutcomeResponse struct {
	Outcome string `json:"outcome"`
}

type ResultResponse struct {
	Bidder string `json:"Bidder"`
	Amount int    `json:"Amount"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run auction_client.go <node URL>")
		return
	}

	nodeURL := os.Args[1]
	reader := bufio.NewReader(os.Stdin)
	var currentBidder string

	fmt.Println("Auction CLI connected to", nodeURL)
	fmt.Println("Commands: login <name>, bid <amount>, result, exit, end")

	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		parts := strings.Split(text, " ")
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "login":
			if len(parts) != 2 {
				fmt.Println("Usage: login <name>")
				continue
			}
			currentBidder = parts[1]
			fmt.Println("Logged in as:", currentBidder)

		case "bid":
			if currentBidder == "" {
				fmt.Println("You must login first.")
				continue
			}
			if len(parts) != 2 {
				fmt.Println("Usage: bid <amount>")
				continue
			}
			var amount int
			_, err := fmt.Sscanf(parts[1], "%d", &amount)
			if err != nil {
				fmt.Println("Invalid amount.")
				continue
			}

			bid := Bid{Bidder: currentBidder, Amount: amount}
			data, _ := json.Marshal(bid)
			resp, err := http.Post(nodeURL+"/bid", "application/json", bytes.NewBuffer(data))
			if err != nil {
				fmt.Println("Error sending bid:", err)
				continue
			}
			var outcome OutcomeResponse
			json.NewDecoder(resp.Body).Decode(&outcome)
			fmt.Println("Bid outcome:", outcome.Outcome)

		case "result":
			resp, err := http.Get(nodeURL + "/result")
			if err != nil {
				fmt.Println("Error getting result:", err)
				continue
			}
			var result ResultResponse
			json.NewDecoder(resp.Body).Decode(&result)
			if result.Bidder == "" {
				fmt.Println("No bids yet.")
			} else {
				fmt.Printf("Highest bid: %s with %d\n", result.Bidder, result.Amount)
			}

		case "end":
			resp, err := http.Post(nodeURL+"/end", "application/json", nil)
			if err != nil {
				fmt.Println("Error ending auction:", err)
				continue
			}
			var result map[string]string
			json.NewDecoder(resp.Body).Decode(&result)
			fmt.Println("Auction status:", result["status"])

		case "exit":
			return

		default:
			fmt.Println("Unknown command.")
		}
	}
}
