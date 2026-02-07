package session

import (
	"database/sql"
	"net/http"
	"numera/config"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

type Session struct {
	*scs.SessionManager
}

func New(db *sql.DB, cfg *config.Config) *Session {
	sm := scs.New()
	sm.Store = sqlite3store.New(db)
	sm.Lifetime = 24 * time.Hour
	sm.IdleTimeout = 20 * time.Minute
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = cfg.IsProd()
	sm.Cookie.Path = "/"

	if cfg.IsProd() {
		sm.Cookie.Name = "__Host-session"
	} else {
		sm.Cookie.Name = "session"
	}

	return &Session{sm}
}

func (s *Session) GetUserID(r *http.Request) int64 {
	return s.GetInt64(r.Context(), "USER_ID")
}

func (s *Session) SetUserID(r *http.Request, userID int64) {
	s.Put(r.Context(), "USER_ID", userID)
}
