package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/airware/vili/redis"
)

// RedisConfig is the redis session configuration
type RedisConfig struct {
	Secure bool
}

type redisService struct {
	config *RedisConfig
}

// InitRedisService initializes the redis session service
func InitRedisService(c *RedisConfig) {
	service = &redisService{
		config: c,
	}
}

func (s *redisService) Login(r *http.Request, w http.ResponseWriter, u *User) error {
	userBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		sessionID := newSessionID()
		success, err := redis.GetClient().SetNX(sessionRedisKey(sessionID), string(userBytes), 0).Result()
		if err != nil {
			return err
		}
		if success {
			http.SetCookie(w, &http.Cookie{
				Name:   sessionCookie,
				Value:  sessionID,
				MaxAge: 60 * 60 * 24, // 1 day
				Path:   "/",
				Secure: s.config.Secure,
			})
			return nil
		}
	}
	return fmt.Errorf("failed to find a unique session ID")
}

func (s *redisService) Logout(r *http.Request, w http.ResponseWriter) error {
	sessionID := getSessionCookie(r)
	err := redis.GetClient().Del(sessionRedisKey(sessionID)).Err()
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		MaxAge: -1, // delete cookie
	})
	return nil
}

func (s *redisService) GetUser(r *http.Request) (*User, error) {
	sessionID := getSessionCookie(r)
	if sessionID == "" {
		return nil, nil
	}
	userBytes, err := redis.GetClient().Get(sessionRedisKey(sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	user := &User{}
	json.Unmarshal(userBytes, user)
	return user, nil
}

func sessionRedisKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
