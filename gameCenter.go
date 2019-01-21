package main

import (
	"context"
	"errors"
	"log"

	ErrorManner "boardgame_server/error"
	pb "boardgame_server/proto"

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

	log.Println(r.State)
}

// 推播遊戲資料by userID
func getGameInfo(userID []int32, gameID int32, gameType string) error {
	r, err := gameCenter.GameInfo(context.Background(), &pb.GameInfoRequest{
		GameID:   gameID,
		GameType: gameType,
		UserID:   userID,
	})

	if err != nil {
		ErrorManner.LogsMessage(err, "GameInfo grpc Error")
		return err
	}

	if r.State == success {
		return nil
	}

	return errors.New("Have no GameInfo success message")
}

// JaipurAction
func jaipurAction(userID int32, gameID int32, gameType string, action JaipurAction) error {
	r, err := gameCenter.JaipurAction(context.Background(), &pb.JaipurActionRequest{
		UserID:   userID,
		GameID:   gameID,
		GameType: gameType,
		Action: &pb.JaipurActionStruct{
			Type:             action.Type,
			Take:             action.Take,
			Sell:             action.Sell,
			SwitchSelfCard:   action.SwitchSelfCard,
			SwitchTargetCard: action.SwitchTargetCard,
		},
	})

	if err != nil {
		ErrorManner.LogsMessage(err, "jaipurAction grpc Error")
		return err
	}

	if r.State == success {
		return nil
	}

	return errors.New("Have no jaipurAction success message")
}

//	call gamecenter to create game
func createGameByGameCenter(gameID int32, gameType string, players Players) error {
	r, err := gameCenter.CreateGame(context.Background(), &pb.CreateGameRequest{
		GameID:   gameID,
		GameType: gameType,
		Players:  compilePlayerToAPIPlayer(players),
	})

	if err != nil {
		ErrorManner.LogsMessage(err, "Create Game grpc Error")
		return err
	}

	if r.State == success {
		return nil
	}

	return errors.New("Have no CreateGame success message")
}

// TODO看可不可以有更好的處理方式，去轉換proto的格式
func compilePlayerToAPIPlayer(players Players) *pb.Players {
	var apiPlayers pb.Players
	for _, player := range players {
		var apiPlayer pb.Player

		apiPlayer.ID = int32(player.ID)
		apiPlayer.Name = player.Name
		apiPlayer.UUID = player.UUID

		apiPlayers.PlayerList = append(apiPlayers.PlayerList, &apiPlayer)
	}

	return &apiPlayers
}
