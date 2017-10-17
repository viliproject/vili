package session

import "net/http"

// HardcodedConfig is the hardcoded session configuration
type HardcodedConfig struct {
	TokenUsers map[string]string
}

type hardcodedService struct {
	config *HardcodedConfig
}

// InitHardcodedService initializes the hardcoded session service
func InitHardcodedService(c *HardcodedConfig) {
	var service Service = &hardcodedService{
		config: c,
	}
	services = append(services, service)
}

func (s *hardcodedService) Login(r *http.Request, w http.ResponseWriter, u *User) (skip bool, err error) {
	return true, nil
}

func (s *hardcodedService) Logout(r *http.Request, w http.ResponseWriter) (skip bool, err error) {
	return true, nil
}

func (s *hardcodedService) GetUser(r *http.Request) (*User, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, nil
	}
	username := s.config.TokenUsers[token]
	if username == "" {
		return nil, nil
	}

	return &User{
		Email:     username,
		Username:  username,
		FirstName: username,
	}, nil
}
