package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/pkg/log"
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using username and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	Login(ctx context.Context, username, password string) (AuthResponse, error)

	Refresh(c context.Context, token string) (AuthResponse, error)
}

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
	IsAdmin() bool
	GetResources() []string
}

type AccountGetter interface {
	Get(ctx context.Context, username, password string) (entity.Account, error)
}

type service struct {
	signingKey       string
	tokenExpiration  int
	rTokenExpiration int
	user             string
	password         string
	accountRepo      AccountGetter
	logger           log.Logger
}

// NewService creates a new authentication service.
func NewService(cfg *config.Config, ar AccountGetter, logger log.Logger) Service {
	return service{
		cfg.JWTSigningKey,
		cfg.JWTExpirationTimeInHours,
		cfg.RefreshTokenExpirationInHours,
		cfg.User,
		cfg.Password,
		ar,
		logger}
}

// Login authenticates a user and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) Login(ctx context.Context, username, password string) (AuthResponse, error) {
	var res AuthResponse
	var err error
	identity := s.authenticate(ctx, username, password)
	if identity == nil {
		return res, errors.Unauthorized("")
	}

	res.AccessToken, err = s.generateJWT(identity)
	if err != nil {
		return res, err
	}
	if s.rTokenExpiration == 0 {
		return res, nil
	}

	res.RefreshToken, err = s.generateRefreshJWT(identity)
	return res, nil
}

// Refresh authenticates a user based on a refresh_token, if it succeeds it will send back
// a new access token.
func (s service) Refresh(ctx context.Context, token string) (AuthResponse, error) {
	var res AuthResponse

	// Extract the id from the token.
	id, err := s.getRefreshJWTSubject(token)
	if err != nil {
		return res, errors.Unauthorized(err.Error())
	}

	// Check the user still exists on the DB
	identity := s.getByID(ctx, id)
	if identity == nil {
		return res, errors.Unauthorized("")
	}

	// Generate an auth token
	res.AccessToken, err = s.generateJWT(identity)
	if err != nil {
		return res, err
	}
	if s.rTokenExpiration == 0 {
		return res, nil
	}

	return res, nil
}

// authenticate authenticates a user using username and password.
// If username and password are correct, an identity is returned. Otherwise, nil is returned.
func (s service) authenticate(ctx context.Context, username, password string) Identity {
	logger := s.logger.With(ctx, "user", username)

	// This is the ENVIRONMENT configured credentials.
	if username == s.user && password == s.password {
		logger.Infof("admin authentication successful")
		return entity.User{
			ID:        "0",
			Name:      s.user,
			Admin:     true,
			Resources: []string{},
		}
	}

	a, err := s.accountRepo.Get(ctx, username, password)
	if err == nil {
		logger.Infof("non-admin authentication successful")
		return entity.User{
			ID:        strconv.Itoa(a.ID),
			Name:      a.UserName,
			Admin:     false,
			Resources: []string{"app1"},
		}
	}

	logger.Infof("authentication failed")
	return nil
}

func (s service) getByID(ctx context.Context, id string) Identity {
	logger := s.logger.With(ctx, "id", id)

	if id == s.user {
		logger.Infof("token refresh successful")
		return entity.User{ID: "100", Name: s.user}
	}

	logger.Infof("token refresh failed")
	return nil
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity Identity) (string, error) {
	// Set custom claims
	claims := &jwtCustomClaims{
		identity.GetID(),
		identity.GetName(),
		identity.IsAdmin(),
		identity.GetResources(),
		jwt.RegisteredClaims{
			Subject:   identity.GetID(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.tokenExpiration))),
		},
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.signingKey))
}

// generateRefreshJWT generates a refresh JWT that encodes an identity.
func (s service) generateRefreshJWT(identity Identity) (string, error) {
	// Set custom claims
	claims := &jwt.RegisteredClaims{
		Subject:   identity.GetName(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.rTokenExpiration))),
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.signingKey))
}

func (s service) getRefreshJWTSubject(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.signingKey), nil
	})

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", err
	}

	return claims.Subject, nil
}
