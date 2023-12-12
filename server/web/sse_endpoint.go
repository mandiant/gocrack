package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mandiant/gocrack/server/authentication"
	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/server/workmgr"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var realtimeTopics = []workmgr.ChannelTopic{workmgr.EngineStatusTopic, workmgr.FinalStatusTopic, workmgr.TaskStatusTopic}

type streamPayload struct {
	Topic   string      `json:"topic"`
	Message interface{} `json:"message"`
}

type connectedUser struct {
	cancel       context.CancelFunc
	data         chan []byte
	claim        *authentication.AuthClaim
	connectedAt  time.Time
	cachedAccess map[string]bool // list of TaskIDs this user is granted access to
	*gin.Context
}

func newConnectedUser(c *gin.Context, cancel context.CancelFunc, user *authentication.AuthClaim) *connectedUser {
	return &connectedUser{
		cancel:       cancel,
		data:         make(chan []byte, 10),
		claim:        user,
		connectedAt:  time.Now().UTC(),
		cachedAccess: make(map[string]bool),
		Context:      c,
	}
}

type RealtimeServer struct {
	users []*connectedUser
	wmgr  *workmgr.WorkerManager
	stor  storage.Backend
	l     sync.Mutex
	hndls []uint
	wg    *sync.WaitGroup
}

func NewRealtimeServer(wmgr *workmgr.WorkerManager, stor storage.Backend) *RealtimeServer {
	rs := &RealtimeServer{
		wmgr:  wmgr,
		stor:  stor,
		users: make([]*connectedUser, 0),
		hndls: make([]uint, len(realtimeTopics)),
		wg:    &sync.WaitGroup{},
	}

	for i, topic := range realtimeTopics {
		hndl, err := wmgr.Subscribe(topic, rs.onManagerMessage)
		if err != nil {
			panic(err)
		}
		log.Debug().Str("topic", string(topic)).Msg("SSE Server subscribed to topic")
		rs.hndls[i] = hndl
	}

	return rs
}

// Stop the Realtime Server by disconnecting all active clients
func (s *RealtimeServer) Stop() {
	s.l.Lock()
	for _, user := range s.users {
		user.cancel()
	}
	s.l.Unlock()

	for _, hndl := range s.hndls {
		s.wmgr.Unsubscribe(hndl)
	}

	s.wg.Wait()
}

func (s *RealtimeServer) onManagerMessage(msg interface{}) {
	var topicName string
	var taskid string

	switch m := msg.(type) {
	case workmgr.TaskEngineStatusBroadcast:
		topicName = "task_engine_status"
		taskid = m.TaskID
	case workmgr.TaskStatusFinalBroadcast:
		topicName = "task_status_final"
		taskid = m.TaskID
	case workmgr.TaskStatusChangeBroadcast:
		topicName = "task_status"
		taskid = m.TaskID
	default:
		return
	}

	// perf: encode the document before putting it into each users queue
	b, err := json.Marshal(&streamPayload{
		Topic:   topicName,
		Message: msg,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode a status message")
		return
	}

	s.l.Lock()
	defer s.l.Unlock()

	for _, user := range s.users {
		// Skip all the checks
		if user.claim.IsAdmin {
			goto SendMessage
		}

		// checked the cached access map to make sure they have access to it
		if canAccess, ok := user.cachedAccess[taskid]; ok {
			if canAccess {
				goto SendMessage
			} else if !canAccess {
				continue
			}
		}

		if !user.claim.IsAdmin {
			canAccess, err := s.stor.CheckEntitlement(user.claim.UserUUID, taskid, storage.EntitlementTask)
			if err != nil {
				if err == storage.ErrNotFound {
					// XXX(cschmitt): Should we cache false here and age it out?
					continue
				}
				log.Error().Err(err).Msg("Failed to check users entitlement against a task")
			}

			if !canAccess {
				user.cachedAccess[taskid] = false
				continue
			}
			user.cachedAccess[taskid] = true
		}

	SendMessage:
		select {
		case user.data <- b:
		default:
			log.Error().
				Err(err).
				Str("remote", user.ClientIP()).
				Str("user", user.claim.UserUUID).
				Msg("User's steam is full. Discarding message")
		}
	}
}

func (s *RealtimeServer) ServeStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	rw := c.Writer
	flusher, ok := rw.(http.Flusher)
	if !ok {
		c.String(500, "Streaming not supported")
		return
	}

	// ping the client
	io.WriteString(rw, ": ping\n\n")
	flusher.Flush()

	claim := getClaimInformation(c)
	log.Debug().
		Str("sock", c.ClientIP()).
		Str("user_uuid", claim.UserUUID).
		Msg("SSE for User Started")

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	user := newConnectedUser(c, cancel, claim)
	s.l.Lock()
	s.users = append(s.users, user)
	s.l.Unlock()

	totalRealtimeConnections.Inc()

	defer func() {
		cancel()
		close(user.data)
		log.Debug().
			Str("sock", c.ClientIP()).
			Str("user_uuid", claim.UserUUID).
			Msg("SSE for User Stopped")

		totalRealtimeConnections.Dec()

		// Remove the user from our tracked users...
		s.l.Lock()
		for i, sock := range s.users {
			if sock == user {
				copy(s.users[i:], s.users[i+1:])
				s.users[len(s.users)-1] = nil
				s.users = s.users[:len(s.users)-1]
			}
		}
		s.l.Unlock()
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case buf, ok := <-user.data:
				if ok {
					totalRealtimeMessages.Inc()
					io.WriteString(rw, "data: ")
					rw.Write(buf)
					io.WriteString(rw, "\n\n")
					flusher.Flush()
				}
			}
		}
	}()

	for {
		select {
		case <-rw.CloseNotify():
			return
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 30):
			io.WriteString(rw, ": ping\n\n")
			flusher.Flush()
		}
	}
}
