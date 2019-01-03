package models

import (
	"context"
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/soooooooon/rock/session"
)

const (
	sessionExpire = time.Hour * 24 // 1 day
)

type Session struct {
	CreatedAt time.Time `json:"created_at"`
	UserId    string    `json:"user_id"`
	Token     string    `json:"token"`
}

func (s *Session) key() string {
	return "rock_session_" + s.Token
}

func (s *Session) Cache(ctx context.Context) error {
	expire := sessionExpire - time.Since(s.CreatedAt)
	if expire <= 0 {
		return errors.New("session expired")
	}

	data, err := jsoniter.Marshal(s)
	if err != nil {
		return err
	}

	r := session.Redis(ctx).Set(s.key(), data, expire)
	return r.Err()
}

func SessionWithToken(ctx context.Context, token string) (*Session, error) {
	s := &Session{Token: token}
	data, err := session.Redis(ctx).Get(s.key()).Bytes()
	if err != nil {
		return nil, err
	}

	if err := jsoniter.Unmarshal(data, s); err != nil {
		return nil, err
	}

	return s, nil
}
