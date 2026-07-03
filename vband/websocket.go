package vband

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	channels          map[string]*Channel
	latestUserID      uint
	ctx               context.Context
	cancel            context.CancelFunc
	server            *http.Server
	logger            *slog.Logger
	upgrader          websocket.Upgrader
	socketUserNameMap map[*websocket.Conn]string
	userIDMap         map[string]uint
}

func NewWebSocketServer() *WebSocketServer {

	upgrader := websocket.Upgrader{
		Subprotocols: []string{"lws-hrs-vband2", "lws-hrs-vband2-ssl"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WebSocketServer{
		channels:          make(map[string]*Channel),
		logger:            slog.Default(),
		upgrader:          upgrader,
		socketUserNameMap: make(map[*websocket.Conn]string),
		userIDMap:         make(map[string]uint),
	}
}

func (s *WebSocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received connection from", "remoteAddr", r.RemoteAddr)
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Failed to read message", "error", err)
			break
		}

		if messageType == websocket.TextMessage {
			stringMsg := string(message)
			stringParts := strings.Split(stringMsg, ",")
			switch stringParts[0] {
			case "CN":
				var connectMsg Connect
				err := connectMsg.Decode(stringMsg)
				if err != nil {
					slog.Error("Failed to decode connect message", "error", err)
					continue
				}
				slog.Info("Received connect message", "name", connectMsg.Name, "deviceID", connectMsg.DeviceID, "clientVersion", connectMsg.ClientVersion)

				existingUserId, exists := s.userIDMap[connectMsg.Name]

				if exists {
					slog.Warn("User already connected", "userName", connectMsg.Name, "userID", existingUserId)
					continue
				}

				s.socketUserNameMap[conn] = connectMsg.Name
				userID := s.GetNextUserID()
				s.userIDMap[connectMsg.Name] = userID
				connectOKMsg := ConnectOK{
					UserID:        userID,
					UserName:      connectMsg.Name,
					ServerVersion: "0.0.0",
				}
				encodedMsg, err := connectOKMsg.Encode()
				if err != nil {
					slog.Error("Failed to encode connect OK message", "error", err)
					continue
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
				if err != nil {
					slog.Error("Failed to send connect OK message", "error", err)
					continue
				}
			case "LC":
				err := conn.WriteMessage(websocket.TextMessage, []byte("CLB"))
				if err != nil {
					slog.Error("Failed to send lobby channel message", "error", err)
					continue
				}
				for _, channel := range s.channels {
					channelListEntry := ChannelListEntry{
						Name:  channel.Name,
						Count: len(channel.Users),
					}
					encodedMsg, err := channelListEntry.Encode()
					if err != nil {
						slog.Error("Failed to encode channel list entry message", "error", err)
						break
					}
					err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
					if err != nil {
						slog.Error("Failed to send channel list entry message", "error", err)
						break
					}
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte("CLC"))
				if err != nil {
					slog.Error("Failed to send channel list complete message", "error", err)
					continue
				}
			case "LU":
				var lobbyUpdateMsg LobbyUpdate
				err := lobbyUpdateMsg.Decode(stringMsg)
				if err != nil {
					slog.Error("Failed to decode lobby update message", "error", err)
					continue
				}
				channel, err := s.GetChannel(lobbyUpdateMsg.Channel)
				if err != nil {
					slog.Error("Failed to get channel for lobby update", "error", err)
					continue
				}
				userListBegin := UserListBegin{Channel: channel.Name}
				encodedMsg, err := userListBegin.Encode()
				if err != nil {
					slog.Error("Failed to encode user list begin message", "error", err)
					continue
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
				if err != nil {
					slog.Error("Failed to send user list begin message", "error", err)
					continue
				}
				for _, user := range channel.Users {
					userListEntry := UserListEntry{
						Channel:  channel.Name,
						UserID:   user.ID,
						UserName: user.Name,
					}
					encodedMsg, err := userListEntry.Encode()
					if err != nil {
						slog.Error("Failed to encode user list entry message", "error", err)
						break
					}
					err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
					if err != nil {
						slog.Error("Failed to send user list entry message", "error", err)
						break
					}
				}
				userListComplete := UserListComplete{Channel: channel.Name}
				encodedMsg, err = userListComplete.Encode()
				if err != nil {
					slog.Error("Failed to encode user list complete message", "error", err)
					continue
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
				if err != nil {
					slog.Error("Failed to send user list complete message", "error", err)
					continue
				}
			case "JC":
				var joinChannelMsg ChannelJoinRequest
				err := joinChannelMsg.Decode(stringMsg)
				if err != nil {
					slog.Error("Failed to decode join channel message", "error", err)
					continue
				}
				channel, err := s.GetChannel(joinChannelMsg.Name)
				if err != nil {
					slog.Error("Failed to get channel for join channel", "error", err)
					continue
				}
				userName, exists := s.socketUserNameMap[conn]
				if !exists {
					slog.Error("User not found for connection")
					continue
				}

				userId, ok := s.userIDMap[userName]
				if !ok {
					slog.Error("User ID not found for user", "userName", userName)
					continue
				}
				user := User{
					ID:   userId,
					Name: userName,
				}
				channel.AddUser(user)
				slog.Info("User joining channel", "userName", userName, "channel", channel.Name)
				channelJoin := ChannelJoin{
					Name:       channel.Name,
					Attributes: "N", //TODO(jwetzell): what is this?
				}

				encodedMsg, err := channelJoin.Encode()
				if err != nil {
					slog.Error("Failed to encode channel join message", "error", err)
					continue
				}
				err = conn.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
				if err != nil {
					slog.Error("Failed to send channel join message", "error", err)
					continue
				}
				channelUpdate := ChannelListUpdate{
					Name:  channel.Name,
					Count: len(channel.Users),
				}
				encodedMsg, err = channelUpdate.Encode()
				if err != nil {
					slog.Error("Failed to encode channel list update message", "error", err)
					continue
				}
				for c := range s.socketUserNameMap {
					err := c.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
					if err != nil {
						slog.Error("Failed to send channel list update message", "error", err)
						continue
					}
				}
			case "SM":
				userName, exists := s.socketUserNameMap[conn]
				if !exists {
					slog.Error("User not found for connection")
					continue
				}
				userId, ok := s.userIDMap[userName]
				if !ok {
					slog.Error("User ID not found for user", "userName", userName)
					continue
				}
				userChannel, err := s.GetUserChannel(userId)
				if err != nil {
					slog.Error("Failed to get user channel", "error", err)
					continue
				}

				var spaceMarkMsg SpaceMark
				err = spaceMarkMsg.Decode(stringMsg)
				if err != nil {
					slog.Error("Failed to decode space/mark message", "error", err)
					continue
				}

				spaceMark := IncomingSpaceMark{
					Channel:  userChannel.Name,
					UserID:   userId,
					UserName: userName,
					Space:    spaceMarkMsg.Space,
					Mark:     spaceMarkMsg.Mark,
				}
				encodedMsg, err := spaceMark.Encode()
				if err != nil {
					slog.Error("Failed to encode space/mark message", "error", err)
					continue
				}
				for _, user := range userChannel.Users {
					if user.ID == userId {
						continue
					}
					for c, name := range s.socketUserNameMap {
						if name == user.Name {
							err := c.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
							if err != nil {
								slog.Error("Failed to send space/mark message", "error", err)
							}
						}
					}
				}
			default:
				slog.Warn("unhandled command", "command", stringParts[0], "message", stringMsg)
			}
		}
	}
	slog.Info("Connection closed", "remoteAddr", r.RemoteAddr)
	username, exists := s.socketUserNameMap[conn]
	if exists {
		delete(s.socketUserNameMap, conn)
		delete(s.userIDMap, username)
		for _, channel := range s.channels {
			if channel.HasUser(s.userIDMap[username]) {
				channel.RemoveUser(s.userIDMap[username])
				channelUpdate := ChannelListUpdate{
					Name:  channel.Name,
					Count: len(channel.Users),
				}
				encodedMsg, err := channelUpdate.Encode()
				if err != nil {
					slog.Error("Failed to encode channel list update message", "error", err)
					continue
				}
				for c := range s.socketUserNameMap {
					err := c.WriteMessage(websocket.TextMessage, []byte(encodedMsg))
					if err != nil {
						slog.Error("Failed to send channel list update message", "error", err)
						continue
					}
				}
			}
		}
	}
}

func (s *WebSocketServer) Start(ctx context.Context, ip string, port uint) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.server = &http.Server{Addr: fmt.Sprintf("%s:%d", ip, port), Handler: s}
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("WebSocket server error", "error", err)
		}
	}()
}

func (s *WebSocketServer) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := s.server.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Error("WebSocket server shutdown error", "error", err)
		}
	}
}

func (s *WebSocketServer) CreateChannel(name string) error {
	_, exists := s.channels[name]
	if exists {
		return fmt.Errorf("channel already exists")
	}
	s.channels[name] = &Channel{Name: name, Users: make(map[uint]User)}
	return nil
}

func (s *WebSocketServer) DeleteChannel(name string) error {
	_, exists := s.channels[name]
	if !exists {
		return fmt.Errorf("channel not found")
	}
	delete(s.channels, name)
	return nil
}

func (s *WebSocketServer) GetChannel(name string) (*Channel, error) {
	channel, exists := s.channels[name]
	if exists {
		return channel, nil
	}
	return nil, fmt.Errorf("channel not found")
}

func (s *WebSocketServer) GetNextUserID() uint {
	userID := s.latestUserID
	s.latestUserID++
	return userID
}

func (s *WebSocketServer) GetUserIDByName(name string) (uint, bool) {
	userID, exists := s.userIDMap[name]
	return userID, exists
}

func (s *WebSocketServer) GetUserChannel(userID uint) (*Channel, error) {
	for _, channel := range s.channels {
		if channel.HasUser(userID) {
			return channel, nil
		}
	}
	return nil, fmt.Errorf("user not found in any channel")
}
