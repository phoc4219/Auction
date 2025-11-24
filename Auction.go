package main

import (
	"fmt"
	"math/rand"
	"sync"
)

// Outcome for bid/result operations
type Outcome string

const (
	Fail      Outcome = "fail"
	Success   Outcome = "success"
	Exception Outcome = "exception"
)

// Bid structure
type Bid struct {
	Bidder string
	Amount int
}

// Auction struct
type Auction struct {
	sync.Mutex
	HighestBid *Bid
	Ended      bool
}

// Auction methods
func (a *Auction) Bid(bidder string, amount int) Outcome {
	a.Lock()
	defer a.Unlock()

	if a.Ended {
		return Fail
	}

	if a.HighestBid == nil || amount > a.HighestBid.Amount {
		a.HighestBid = &Bid{Bidder: bidder, Amount: amount}
		return Success
	}
	return Fail
}

func (a *Auction) Result() *Bid {
	a.Lock()
	defer a.Unlock()
	return a.HighestBid
}

func (a *Auction) EndAuction() {
	a.Lock()
	defer a.Unlock()
	a.Ended = true
}

// AuctioneerNode represents a replicated auctioneer node
type AuctioneerNode struct {
	ID      int
	Auction *Auction
	Peers   []*AuctioneerNode
}

// Bid on this node and replicate to peers
func (n *AuctioneerNode) Bid(bidder string, amount int) Outcome {
	outcome := n.Auction.Bid(bidder, amount)

	var wg sync.WaitGroup
	for _, peer := range n.Peers {
		wg.Add(1)
		go func(p *AuctioneerNode) {
			defer wg.Done()
			p.Auction.Bid(bidder, amount)
		}(peer)
	}
	wg.Wait()

	return outcome
}

// Query the result from this node
func (n *AuctioneerNode) Result() *Bid {
	return n.Auction.Result()
}

// End the auction on all nodes
func (n *AuctioneerNode) EndAuction() {
	n.Auction.EndAuction()
	for _, peer := range n.Peers {
		peer.Auction.EndAuction()
	}
}

func pickRandomNode(nodes []*AuctioneerNode) *AuctioneerNode {
	alive := []*AuctioneerNode{}
	for _, n := range nodes {
		if n != nil {
			alive = append(alive, n)
		}
	}
	if len(alive) == 0 {
		return nil
	}
	return alive[rand.Intn(len(alive))]
}

// Demo
func main() {
	// Create 3 auctioneer nodes
	node1 := &AuctioneerNode{ID: 1, Auction: &Auction{}}
	node2 := &AuctioneerNode{ID: 2, Auction: &Auction{}}
	node3 := &AuctioneerNode{ID: 3, Auction: &Auction{}}

	nodes := []*AuctioneerNode{node1, node2, node3}
	// Set peers (full mesh)
	node1.Peers = []*AuctioneerNode{node2, node3}
	node2.Peers = []*AuctioneerNode{node1, node3}
	node3.Peers = []*AuctioneerNode{node1, node2}

	var currentBidder string = ""

	fmt.Println("Auction started. Commands:")
	fmt.Println("  login <name>")
	fmt.Println("  bid <amount>")
	fmt.Println("  result")
	fmt.Println("  end")

	var cmd string

	for {
		fmt.Print("> ")
		_, err := fmt.Scan(&cmd)
		if err != nil {
			continue
		}

		switch cmd {
		case "login":
			var name string
			_, err := fmt.Scan(&name)
			if err != nil {
				fmt.Println("Usage: login <name>")
				continue
			}
			currentBidder = name
			fmt.Println("Logged in as:", currentBidder)

		case "bid":
			if currentBidder == "" {
				fmt.Println("You must login first. Use: login <name>")
				continue
			}

			var amount int
			_, err := fmt.Scan(&amount)
			if err != nil {
				fmt.Println("Invalid bid command. Usage: bid <amount>")
				continue
			}

			node := pickRandomNode(nodes)
			if node == nil {
				fmt.Println("No auctioneer nodes available!")
				continue
			}
			outcome := node.Bid(currentBidder, amount)

			fmt.Println("bid Outcome:", outcome)

		case "result":
			node := pickRandomNode(nodes)
			if node == nil {
				fmt.Println("No nodes alive.")
				continue
			}
			b := node.Result()
			if b == nil {
				fmt.Println("No bids yet.")
			} else {
				fmt.Printf("Highest bid: %s with %d\n", b.Bidder, b.Amount)
			}

		case "end":
			node := pickRandomNode(nodes)
			if node == nil {
				fmt.Println("Cannot end auction — no nodes alive.")
				continue
			}
			node.EndAuction()

			b := node.Result()
			if b == nil {
				fmt.Println("No bids yet.")
			} else {
				fmt.Printf("Highest bid: %s with %d\n", b.Bidder, b.Amount)
			}

			fmt.Println("Auction has finished.")
			return

		default:
			fmt.Println("Error: Unknown command.")
		}
	}
}
