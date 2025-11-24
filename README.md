# Auction
Mandatory Activities 5

How to run the Auction
1. Open several terminals navigate to the program root folder
2. go run auction_node.go 8001 http://localhost:8002 http://locahost:8003
3. go run auction_node.go 8002 http://localhost:8001 http://locahost:8003
4. go run auction_node.go 8003 http://localhost:8001 http://locahost:8002
5. then run the client in a new terminal
6. go run auction_client.go http:/localhost:8001(or any of the operating nodes localhost)
7. You're now Auctioning.
