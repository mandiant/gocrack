package notifications

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/rs/zerolog/log"
	gomail "gopkg.in/gomail.v2"
)

type userRecord struct {
	UserUUID string
	Email    string
}

// Config describes the available configuration options for the email notification engine
type Config struct {
	EmailServer struct {
		Address         string  `yaml:"address"`
		Port            int     `yaml:"port"`
		Username        string  `yaml:"username"`
		Password        string  `yaml:"password"`
		SkipInvalidCert bool    `yaml:"skip_invalid_cert"`
		Certifcate      *string `yaml:"certificate"`
		ServerName      *string `yaml:"server_name"`
	} `yaml:"email_server"`
	Enabled       bool   `yaml:"enabled"`
	FromAddress   string `yaml:"from_address"`
	PublicAddress string `yaml:"public_address"`
}

func (s *Config) Validate() error {
	if !s.Enabled {
		return nil
	}

	if s.EmailServer.Address == "" {
		return errors.New("notifications.email_server.address must not be empty")
	}

	if s.FromAddress == "" {
		return errors.New("notifications.from_address must not be empty")
	}

	if s.PublicAddress == "" {
		return errors.New("notifications.public_address must not be empty. It must point to the URL where the UI is")
	}

	if s.EmailServer.Port == 0 {
		s.EmailServer.Port = 25
	}

	return nil
}

// Engine is a context that provides notification capabilities via email
type Engine struct {
	cfg    Config
	stor   storage.Backend
	dialer *gomail.Dialer
	cache  *CacheLastPasswordSent
	mu     *sync.Mutex
}

// New builds the notification engine
func New(cfg Config, stor storage.Backend) (*Engine, error) {
	dialer := gomail.NewDialer(cfg.EmailServer.Address, cfg.EmailServer.Port, cfg.EmailServer.Username, cfg.EmailServer.Password)
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: cfg.EmailServer.SkipInvalidCert,
	}

	if cfg.EmailServer.Certifcate != nil && *cfg.EmailServer.Certifcate != "" {
		certp := x509.NewCertPool()
		if ok := certp.AppendCertsFromPEM([]byte(*cfg.EmailServer.Certifcate)); !ok {
			return nil, errors.New("failed to build cert pool with ca certificate")
		}
		dialer.TLSConfig.RootCAs = certp

		if cfg.EmailServer.ServerName != nil && *cfg.EmailServer.ServerName != "" {
			dialer.TLSConfig.ServerName = *cfg.EmailServer.ServerName
		} else {
			dialer.TLSConfig.ServerName = cfg.EmailServer.Address
		}
	}

	return &Engine{
		cfg:    cfg,
		stor:   stor,
		dialer: dialer,
		cache:  NewCache(10 * time.Minute),
		mu:     &sync.Mutex{},
	}, nil
}

// Stop the notification engine
func (s *Engine) Stop() {
	s.cache.Stop()
}

// lookupUsersForNotification takes a task ID, looks up the granted entitlements for it, and returns a list of users
// along with their email address.
func (s *Engine) lookupUsersForNotification(taskID string) ([]userRecord, error) {
	// get all the users who are entitled to this document
	entitlements, err := s.stor.GetEntitlementsForTask(taskID)
	if err != nil {
		return []userRecord{}, err
	}

	records := make([]userRecord, 0)
	for _, ent := range entitlements {
		user, err := s.stor.GetUserByID(ent.UserUUID)
		// Unable to get users or their email address hasnt been set yet, lets skip it
		if err != nil || user.EmailAddress == "" {
			continue
		}

		records = append(records, userRecord{
			UserUUID: ent.UserUUID,
			Email:    user.EmailAddress,
		})
	}
	return records, nil
}

// CrackedPassword is called whenever a password has been cracked and will send out a notification email
// if one can be sent.
func (s *Engine) CrackedPassword(taskID string) error {
	// first ensure the task is there?
	task, err := s.stor.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	// We have to acquire a lock here because if passwords crack in quick succession,
	// we could spam users since this will be called from a short lived goroutine
	s.mu.Lock()
	// Block the email from going out
	if canSend := s.cache.CanSendEmail(taskID); !canSend {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	users, err := s.lookupUsersForNotification(taskID)
	if err != nil {
		return err
	}

	mails := make([]*gomail.Message, len(users))
	for i, user := range users {
		m := gomail.NewMessage()
		notificationsSent.WithLabelValues("new_passwords").Inc()

		m.SetHeader("From", s.cfg.FromAddress)
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", fmt.Sprintf("New Password(s) for %s", taskID))
		m.AddAlternativeWriter("text/html", func(w io.Writer) error {
			template := templateCrackedPassword{
				TaskName:  task.TaskName,
				TaskID:    taskID,
				PublicURL: s.cfg.PublicAddress,
			}

			if task.CaseCode != nil {
				template.CaseCode = *task.CaseCode
			}
			return emailCrackedTemplate.Execute(w, template)
		})

		if e := log.Debug(); e.Enabled() {
			e.Str("to", user.Email).Str("task_id", taskID).Str("type", "cracked_passwords").Int("id", i).Msg("Generated email")
		}
		mails[i] = m
	}

	return s.dialer.DialAndSend(mails...)
}

// TaskStatusChanged is called whenever a task status has changed
func (s *Engine) TaskStatusChanged(taskID string, newStatus storage.TaskStatus) error {

	// discard emails on useless status changes
	switch newStatus {
	case storage.TaskStatusDequeued, storage.TaskStatusQueued:
		return nil
	default:
	}

	// first ensure the task is there?
	task, err := s.stor.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	users, err := s.lookupUsersForNotification(taskID)
	if err != nil {
		return err
	}

	mails := make([]*gomail.Message, len(users))
	for i, user := range users {
		m := gomail.NewMessage()
		notificationsSent.WithLabelValues("task_status_changed").Inc()
		m.SetHeader("From", s.cfg.FromAddress)
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", fmt.Sprintf("Task Status change for %s", taskID))
		m.AddAlternativeWriter("text/html", func(w io.Writer) error {
			template := templateTaskStatusChanged{
				TaskID:    taskID,
				NewStatus: string(newStatus),
				TaskName:  task.TaskName,
				PublicURL: s.cfg.PublicAddress,
			}
			if task.CaseCode != nil {
				template.CaseCode = *task.CaseCode
			}
			return emailStatusChanged.Execute(w, template)
		})

		if e := log.Debug(); e.Enabled() {
			e.Str("to", user.Email).Str("task_id", taskID).Str("type", "task_status_changed").Int("id", i).Msg("Generated email")
		}
		mails[i] = m
	}

	return s.dialer.DialAndSend(mails...)
}
