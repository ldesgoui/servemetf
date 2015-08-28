package models

import (
	"net/rpc"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

type ServerBootstrap struct {
	LobbyId       uint
	Info          ServerRecord
	Players       []string
	BannedPlayers []string
}

type Args struct {
	Id      uint
	Info    ServerRecord
	Type    LobbyType
	League  string
	Map     string
	SteamId string
}

var Pauling *rpc.Client

type Event map[string]interface{}

func PaulingConnect() {
	helpers.Logger.Debug("Connecting to Pauling on port %s", config.Constants.PaulingPort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.PaulingPort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	Pauling = client
	helpers.Logger.Debug("Connected!")
}

func AllowPlayer(lobbyId uint, steamId string) error {
	return Pauling.Call("Pauling.AllowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func DisallowPlayer(lobbyId uint, steamId string) error {
	return Pauling.Call("Pauling.DisallowPlayer", &Args{Id: lobbyId, SteamId: steamId}, &Args{})
}

func SetupServer(lobbyId uint, info ServerRecord, lobbyType LobbyType, league string,
	mapName string) error {

	args := &Args{
		Id:     lobbyId,
		Info:   info,
		Type:   lobbyType,
		League: league,
		Map:    mapName}
	return Pauling.Call("Pauling.SetupServer", args, &Args{})
}

func End(lobbyId uint) {
	Pauling.Call("Pauling.End", &Args{Id: lobbyId}, &Args{})
}
