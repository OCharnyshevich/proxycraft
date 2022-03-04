package main

import (
	"fmt"
	"github.com/OCharnyshevich/proxycraft/proxy"
	"github.com/OCharnyshevich/proxycraft/proxy/network"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	mcNet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"log"
)

func main() {
	color.NoColor = false
	config := &proxy.NewConfig

	p, err := proxy.New(config)
	if err != nil {
		panic(err)
	}

	network.EventsListener{
		GameStart:      onGameStart,
		ChatMsg:        onChatMsg,
		KickDisconnect: onKickDisconnect,
		Death:          onDeath,
	}.Attach(p.Network())

	events := p.Network().Events().(*network.Events)
	events.AddListener(
		network.PacketHandler{Priority: 64, ID: packetid.Camera, F: func(client *mcNet.Conn, server *mcNet.Conn, packet pk.Packet) error {
			p.Logging().InfoF("Hit")
			return nil
		}},
	)
	events.AddGeneric(
		network.PacketHandler{Priority: 64, F: func(client *mcNet.Conn, server *mcNet.Conn, packet pk.Packet) error {

			switch packet.ID {
			case 0,
				packetid.EntityHeadRotation,
				packetid.EntityVelocity,
				packetid.UpdateJigsawBlock,
				packetid.UpdateStructureBlock,
				packetid.DestroyEntity,
				packetid.EntityMetadata,
				packetid.KeepAliveClientbound,
				packetid.MapChunk,
				packetid.UpdateLight,
				packetid.EntityUpdateAttributes,
				packetid.BlockChange,
				packetid.SpawnEntityLiving,
				packetid.EntityLook,
				packetid.UnloadChunk,
				packetid.UpdateViewPosition,
				packetid.PlayerInfo,
				packetid.ChatClientbound,
				packetid.MultiBlockChange,
				packetid.CustomPayloadClientbound,
				packetid.PositionClientbound,
				packetid.EntityTeleport:
			case packetid.OpenWindow:
				p.Logging().InfoF("OpenWindow")
			case packetid.WindowItems:
				p.Logging().InfoF("WindowItems")
			case packetid.CloseWindowClientbound:
				p.Logging().InfoF("CloseWindowClientbound")
			case packetid.SetSlot:
				p.Logging().InfoF("SetSlot")
			case packetid.Animation:
				p.Logging().InfoF("Animation")
			case packetid.PickItem:
				p.Logging().InfoF("PickItem")
			case packetid.UpdateTime:
				var (
					wordAge   pk.Long
					timeOfDay pk.Long
				)
				_ = packet.Scan(&wordAge, &timeOfDay)
				//p.Logging().InfoF("UpdateTime; Word age: %v, time of day: %v", wordAge, timeOfDay)
				p.Broadcast(fmt.Sprintf("UpdateTime; Word age: %v, time of day: %v", wordAge, timeOfDay))
				err := client.WritePacket(pk.Marshal(
					packetid.UpdateTime,
					pk.Long(0), pk.Long(0),
				))
				if err != nil {
					p.Logging().Fail(err)
				}
			case packetid.GameStateChange:
				var (
					reason pk.VarInt
					value  pk.Float
				)
				_ = packet.Scan(&reason, &value)
				p.Logging().InfoF("GameStateChange; Reason: %d, value: %f", reason, value)
			case packetid.SoundEffect:
				var (
					id       pk.VarInt
					category pk.VarInt
					entityId pk.VarInt
					volume   pk.Float
					pitch    pk.Float
				)
				_ = packet.Scan(&id, &category, &entityId, &volume, &pitch)
				p.Logging().InfoF("SoundEffect; Id: %d | bookOpen: %d | filterActive: %d | volume: %f | pitch: %f", id, category, entityId, volume, pitch)
			case packetid.BlockAction:
				p.Logging().InfoF("BlockAction")
				err := client.WritePacket(pk.Marshal(
					packetid.UpdateTime,
					pk.Long(0), pk.Long(0),
				))
				if err != nil {
					p.Logging().Fail(err)
				}
			case packetid.UpdateHealth:
				var (
					health         pk.Float
					food           pk.VarInt
					foodSaturation pk.Float
				)
				_ = packet.Scan(&health, &food, &foodSaturation)
				p.Logging().InfoF("UpdateHealth; Health: %.0f, food: %d, saturation: %.0f", health, food, foodSaturation)
			default:
				p.Logging().InfoF("Read packet: 0x%X", packet.ID)
			}
			return nil
		}},
	)

	p.Load()
}

func onDeath() error {
	log.Println("Died and Respawned")
	// If we exclude Respawn(...) then the player won't press the "Respawn" button upon death
	return nil
}

func onGameStart() error {
	log.Println("Game start")

	return nil
}

func onChatMsg(c chat.Message, pos byte, uuid uuid.UUID) error {
	log.Println("Chat:", c, pos, uuid)
	return nil
}

func onKickDisconnect(c chat.Message) error {
	log.Println("KickDisconnect:", c)
	return nil
}
