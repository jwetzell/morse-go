package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwetzell/morse-go/kob"
)

var ip string
var port int
var debug bool

func init() {
	flag.StringVar(&ip, "ip", "0.0.0.0", "IP address to listen on")
	flag.IntVar(&port, "port", 7890, "port to listen on")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
}

type Server struct {
	wires               map[uint16]*kob.Wire
	stations            map[string]*kob.Station
	addrWireConnections map[netip.AddrPort]uint16
	addrStationMap      map[netip.AddrPort]string
	conn                *net.UDPConn
	ctx                 context.Context
}

func main() {
	flag.Parse()

	if ip == "" {
		slog.Error("IP address is required")
		return
	}

	if port <= 0 || port > 65535 {
		slog.Error("invalid port number", "port", port)
		return
	}

	localAddr := net.ParseIP(ip)
	if localAddr == nil {
		slog.Error("invalid IP address", "ip", ip)
		return
	}

	var logLevel slog.Level = slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: localAddr, Port: port})
	if err != nil {
		slog.Error("failed to start UDP server", "error", err)
		return
	}
	slog.Info("UDP server started", "ip", ip, "port", port)

	defer conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := &Server{
		wires:               make(map[uint16]*kob.Wire),
		addrWireConnections: make(map[netip.AddrPort]uint16),
		addrStationMap:      make(map[netip.AddrPort]string),
		stations:            make(map[string]*kob.Station),
		conn:                conn,
		ctx:                 ctx,
	}

	go func() {
		<-server.ctx.Done()
		slog.Info("Shutting down server...")
		server.conn.Close()
	}()

	buf := make([]byte, 1024)
	for {
		n, addr, err := server.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-server.ctx.Done():
				slog.Info("server stopped")
				return
			default:
				slog.Error("error reading from UDP", "error", err)
				continue
			}
		}

		server.handleData(buf[:n], addr)
	}
}

func (s *Server) handleData(data []byte, addr *net.UDPAddr) {
	if len(data) < 2 {
		slog.Warn("received packet too short to contain a valid header", "from", addr)
		return
	}

	packetType := data[0]
	switch packetType {
	case 0x02:
		slog.Info("received DisconnectPacket", "from", addr)
		stationID, exists := s.addrStationMap[addr.AddrPort()]
		if !exists {
			slog.Warn("received DisconnectPacket from unknown station", "from", addr)
			return
		}
		wireConnection, exists := s.addrWireConnections[addr.AddrPort()]
		if !exists {
			slog.Warn("received DisconnectPacket from address without a connection", "from", addr)
			return
		}
		wire, exists := s.wires[wireConnection]
		if !exists {
			slog.Warn("wire connection from address does not exist", "from", addr, "wireConnection", wireConnection)
			return
		}

		station, exists := s.stations[stationID]
		if !exists {
			slog.Warn("station from address does not exist", "from", addr, "stationID", stationID)
			return
		}

		wire.Disconnect(station)
		slog.Info("disconnected station from wire", "stationID", stationID, "wire", wireConnection)
		delete(s.stations, stationID)
		delete(s.addrStationMap, addr.AddrPort())
	case 0x3:
		var dataPacket kob.DataPacket
		err := dataPacket.UnmarshalBinary(data)
		if err != nil {
			slog.Error("failed to unmarshal DataPacket", "error", err)
			return
		}
		if len(dataPacket.CodeList) == 0 {
			var idPacket kob.IDPacket
			err := idPacket.UnmarshalBinary(data)
			if err != nil {
				slog.Error("failed to unmarshal IDPacket", "error", err)
				return
			}
			wireConnection, exists := s.addrWireConnections[addr.AddrPort()]
			if !exists {
				slog.Warn("received IDPacket from address without a connection", "address", addr)
				return
			}
			wire, exists := s.wires[wireConnection]
			if !exists {
				slog.Warn("wire connection from address does not exist", "address", addr, "wireConnection", wireConnection)
				return
			}
			station := kob.NewStation(s.ctx, idPacket.StationID, idPacket.Version)
			wire.Connect(station)
			s.stations[station.ID()] = station
			s.addrStationMap[addr.AddrPort()] = station.ID()
		} else {
			station, exists := s.stations[dataPacket.StationID]
			if !exists {
				slog.Warn("received DataPacket for unknown station", "stationID", dataPacket.StationID)
				return
			}
			station.PushCodeList(dataPacket.CodeList)

			wire, exists := s.addrWireConnections[addr.AddrPort()]
			if !exists {
				slog.Warn("received DataPacket from address without a connection", "address", addr)
				return
			}

			for addrPort, wireNo := range s.addrWireConnections {
				if wireNo == wire && addrPort != addr.AddrPort() {
					_, err := s.conn.WriteToUDPAddrPort(data, addrPort)
					if err != nil {
						slog.Error("failed to write DataPacket to address", "address", addrPort, "error", err)
					}
				}
			}

		}
	case 0x4:
		var connectPacket kob.ConnectPacket
		err := connectPacket.UnmarshalBinary(data)
		if err != nil {
			slog.Error("failed to unmarshal ConnectPacket", "error", err)
			return
		}
		if _, exists := s.wires[connectPacket.Wire]; !exists {
			slog.Info("setting up wire", "wire", connectPacket.Wire)
			s.wires[connectPacket.Wire] = kob.NewWire(s.ctx)
		}

		existingConnection, exists := s.addrWireConnections[addr.AddrPort()]
		if exists && existingConnection != connectPacket.Wire {
			slog.Warn("address already connected to a different wire, overwriting connection", "address", addr, "existingWire", existingConnection, "newWire", connectPacket.Wire)
		}
		s.addrWireConnections[addr.AddrPort()] = connectPacket.Wire

		ackPacket := kob.AckPacket{}
		ackData, err := ackPacket.MarshalBinary()
		if err != nil {
			slog.Error("failed to marshal AckPacket", "error", err)
			return
		}
		_, err = s.conn.WriteToUDP(ackData, addr)
		if err != nil {
			slog.Error("failed to write AckPacket to address", "address", addr, "error", err)
			return
		}
	case 0x5:
		slog.Debug("received AckPacket")
	default:
		slog.Warn("received unknown packet type", "type", packetType, "from", addr)
	}
}
