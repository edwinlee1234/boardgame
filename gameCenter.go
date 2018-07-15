package main

import (
	"context"
	"fmt"
	"log"

	pb "./proto"

	"google.golang.org/grpc"
)

// 連接gamecenter的grpc
func connectGameCenter() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(gameCenterAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	gameCenter = pb.NewGameCenterClient(conn)

	// test connect
	r, err := gameCenter.Ping(context.Background(), &pb.TestRequest{})
	if err != nil {
		log.Fatalf("could not cennect gamecenter: %v", err)
	}

	fmt.Println(r.State)
}

//	call gamecenter to create game
func createGameByGameCenter(gameID int, gameType string) {
	r, err := gameCenter.CreateGame(context.Background(), &pb.CreateGameRequest{
		GameID:   int32(gameID),
		GameType: gameType,
	})

	if err != nil {
		log.Fatalf("create game error by gamecenter: %v", err)
	}

	fmt.Println(r.State)
}
