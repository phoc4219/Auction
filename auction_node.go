package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Bid struct {
	Bidder string `json:"bidder"`
	Amount int    `json:"amount"`
}

type Outcome string

const (
	Fail      Outcome = "fail"
	Success   Outcome = "success"
	Exception Outcome = "exception"
)

type Auction struct {
	HighestBid *Bid
	Ended      bool
}

func (a *Auction) Bid(bidder string, amount int) Outcome {
	if a.Ended {
		return Fail
	}
	if a.HighestBid == nil || amount > a.HighestBid.Amount && amount != a.HighestBid.Amount {
		a.HighestBid = &Bid{Bidder: bidder, Amount: amount}
		return Success
	}
	return Fail
}

type Node struct {
	Auction *Auction
	Peers   []string
	ID      string
}

func (n *Node) BidHandler(w http.ResponseWriter, r *http.Request) {
	var bid Bid
	json.NewDecoder(r.Body).Decode(&bid)
	outcome := n.Auction.Bid(bid.Bidder, bid.Amount)
	// replicate sequentially
	for _, peer := range n.Peers {
		http.Post(peer+"/bid", "application/json", bytes.NewBuffer(mustJSON(bid)))
	}
	json.NewEncoder(w).Encode(map[string]string{"outcome": string(outcome)})
}

func (n *Node) ResultHandler(w http.ResponseWriter, r *http.Request) {
	if n.Auction.HighestBid == nil {
		json.NewEncoder(w).Encode(map[string]string{"bidder": "", "amount": "0"})
	} else {
		json.NewEncoder(w).Encode(n.Auction.HighestBid)
	}
}

func (n *Node) EndAuctionHandler(w http.ResponseWriter, r *http.Request) {
	if n.Auction.Ended {
		w.Write([]byte(`{"status":"already ended"}`))
		return
	}
	n.Auction.Ended = true

	// replicate sequentially to peers
	for _, peer := range n.Peers {
		http.Post(peer+"/end", "application/json", nil)
	}

	w.Write([]byte(`{"status":"ended"}`))
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port> [peer1 peer2 ...]")
		return
	}
	port := os.Args[1]
	peers := os.Args[2:]

	node := &Node{
		Auction: &Auction{},
		Peers:   peers,
		ID:      port,
	}

	http.HandleFunc("/bid", node.BidHandler)
	http.HandleFunc("/result", node.ResultHandler)
	http.HandleFunc("/end", node.EndAuctionHandler)

	fmt.Println("Node running on port", port)
	http.ListenAndServe(":"+port, nil)
}
