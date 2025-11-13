package main

import (
	"fmt"
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

// Demo
func main() {
	// Create 3 auctioneer nodes
	node1 := &AuctioneerNode{ID: 1, Auction: &Auction{}}
	node2 := &AuctioneerNode{ID: 2, Auction: &Auction{}}
	node3 := &AuctioneerNode{ID: 3, Auction: &Auction{}}

	// Set peers (full mesh)
	node1.Peers = []*AuctioneerNode{node2, node3}
	node2.Peers = []*AuctioneerNode{node1, node3}
	node3.Peers = []*AuctioneerNode{node1, node2}

	// Bidding through any node
	fmt.Println("Bidder1 bids 100:", node1.Bid("Bidder1", 100))
	fmt.Println("Bidder2 bids 90:", node2.Bid("Bidder2", 90))
	fmt.Println("Bidder2 bids 150:", node3.Bid("Bidder2", 150))

	// End auction
	node1.EndAuction()

	// Query result from any node
	fmt.Println("Auction result from node1:", node1.Result())
	fmt.Println("Auction result from node2:", node2.Result())
	fmt.Println("Auction result from node3:", node3.Result())
}
