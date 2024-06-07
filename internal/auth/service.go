package auth

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/app"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
)

// Service encapsulates the authentication logic.
type Service interface {
	// authenticate authenticates a user using username and password.
	// It returns a JWT token if authentication succeeds. Otherwise, an error is returned.
	Login(ctx context.Context, username, password string) (LoginResponse, error)

	Refresh(c context.Context, token string) (LoginResponse, error)
}

type AccountGetter interface {
	Get(ctx context.Context, username, password string) (entity.Account, error)
	GetByUsername(ctx context.Context, username string) (entity.Account, error)
}

type service struct {
	signingKey       string
	tokenExpiration  int
	rTokenExpiration int
	user             string
	password         string
	accountRepo      AccountGetter
	appRepo          app.Repository
	logger           log.Logger
}

// NewService creates a new authentication service.
func NewService(cfg *config.Config, ar AccountGetter, appRepo app.Repository, logger log.Logger) Service {
	return service{
		cfg.JWTSigningKey,
		cfg.JWTExpirationTimeInHours,
		cfg.RefreshTokenExpirationInHours,
		cfg.User,
		cfg.Password,
		ar,
		appRepo,
		logger}
}

// Login authenticates a user and generates a JWT token if authentication succeeds.
// Otherwise, an error is returned.
func (s service) Login(ctx context.Context, username, password string) (LoginResponse, error) {
	var res LoginResponse
	var err error
	identity := s.authenticate(ctx, username, password)
	if identity == nil {
		s.logger.With(ctx).Infof("account %s does not exist", username)
		return res, errors.Unauthorized("account does not exist")
	}

	res.AccessToken, err = s.generateJWT(identity)
	if err != nil {
		s.logger.With(ctx).Infof("error generating the jwt token %v", err)
		return res, err
	}
	if s.rTokenExpiration == 0 {
		s.logger.With(ctx).Info("token expiration is 0")
		return res, nil
	}

	res.RefreshToken, err = s.generateRefreshJWT(identity)
	return res, nil
}

// Refresh authenticates a user based on a refresh_token, if it succeeds it will send back
// a new access token.
func (s service) Refresh(ctx context.Context, token string) (LoginResponse, error) {
	var res LoginResponse

	// Extract the id from the token.
	id, err := s.getRefreshJWTSubject(token)
	if err != nil {
		s.logger.With(ctx).Infof("problem extracting the user id from the input token - %v", err)
		return res, errors.Unauthorized(err.Error())
	}

	// Check the user still exists on the DB
	identity := s.getByID(ctx, id)
	if identity == nil {
		s.logger.With(ctx).Infof("error retrieving user %s by id", id)
		return res, errors.Unauthorized("account not found")
	}

	// Generate an auth token
	res.AccessToken, err = s.generateJWT(identity)
	if err != nil {
		s.logger.With(ctx).Infof("error generating the access token - %v", err)
		return res, err
	}

	if s.rTokenExpiration == 0 {
		s.logger.With(ctx).Info("token expiration is 0")
		return res, nil
	}

	return res, nil
}

// authenticate authenticates a user using username and password.
// If username and password are correct, an identity is returned. Otherwise, nil is returned.
func (s service) authenticate(ctx context.Context, username, password string) acl.Identity {
	logger := s.logger.With(ctx, "user", username)

	// This is the ENVIRONMENT configured credentials.
	if username == s.user && password == s.password {
		logger.Debug("admin authentication successful")
		return entity.User{
			ID:                     "0",
			Name:                   s.user,
			Admin:                  true,
			RequiresPasswordChange: false,
			Resources:              []string{},
		}
	}

	a, err := s.accountRepo.Get(ctx, username, password)
	if err == nil {
		logger.Debug("non-admin authentication successful")
		u := entity.User{
			ID:                     strconv.Itoa(a.ID),
			Name:                   a.UserName,
			Admin:                  false,
			RequiresPasswordChange: (a.RequiresPasswordChange == 1),
			Resources:              a.GetResources(),
		}
		return u
	}

	logger.Infof("authentication failed")
	return nil
}

func (s service) getByID(ctx context.Context, id string) acl.Identity {
	logger := s.logger.With(ctx, "id", id)

	// Look up for the default configured user?
	if id == s.user {
		return entity.User{
			ID:   "0",
			Name: s.user,
		}
	}

	// Look up for accounts on the database
	account, err := s.accountRepo.GetByUsername(ctx, id)
	if err != nil {
		logger.With(ctx).Infof("account %s does not exist", id)
		return nil
	}

	return entity.User{
		ID:   strconv.Itoa(account.ID),
		Name: account.UserName,
	}
}

// generateJWT generates a JWT that encodes an identity.
func (s service) generateJWT(identity acl.Identity) (string, error) {
	return acl.GenerateJWTToken(identity, 0, s.signingKey, s.tokenExpiration)
}

// generateRefreshJWT generates a refresh JWT that encodes an identity.
func (s service) generateRefreshJWT(identity acl.Identity) (string, error) {
	return acl.GenerateRefreshToken(identity, s.signingKey, s.rTokenExpiration)
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
