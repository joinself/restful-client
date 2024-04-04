package acl

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// Identity represents an authenticated user identity.
type Identity interface {
	// GetID returns the user ID.
	GetID() string
	// GetName returns the user name.
	GetName() string
	IsAdmin() bool
	GetResources() []string
	IsPasswordChangeRequired() bool
}

type JWTCustomClaims struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Admin                    bool     `json:"admin"`
	Resources                []string `json:"resources"`
	Token                    int      `json:"tid"`
	IsPasswordChangeRequired bool     `json:"change_password"`
	jwt.RegisteredClaims
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(c echo.Context) Identity {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := token.Claims.(*JWTCustomClaims) // by default claims is of type `jwt.MapClaims`
	if !ok {
		return nil
	}
	return entity.User{
		ID:                     claims.ID,
		Name:                   claims.Name,
		Admin:                  claims.Admin,
		Resources:              claims.Resources,
		RequiresPasswordChange: claims.IsPasswordChangeRequired,
	}
}

func CurrentToken(c echo.Context) (int, bool) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return 0, false
	}
	claims, ok := token.Claims.(*JWTCustomClaims) // by default claims is of type `jwt.MapClaims`
	if !ok {
		return 0, false
	}
	return claims.Token, true
}

// HasAccessToResource checks if the current user has access to a specific resource.
func HasAccessToResource(c echo.Context, resource string) bool {
	u := CurrentUser(c)
	if u == nil {
		c.JSON(http.StatusNotFound, response.Error{
			Status:  http.StatusNotFound,
			Error:   "Not found",
			Details: "The requested resource does not exist, or you don't have permissions to access it",
		})
		return false
	}

	if u.IsAdmin() {
		return true
	}

	if u.IsPasswordChangeRequired() {
		c.JSON(http.StatusLocked, response.Error{
			Status:  http.StatusLocked,
			Error:   "You're required to change your password",
			Details: "Please change your password before consuming the api.",
		})

		return false
	}

	if isAPermittedResource(u.GetResources(), resource) {
		return true
	}

	c.JSON(http.StatusNotFound, response.Error{
		Status:  http.StatusNotFound,
		Error:   "Not found",
		Details: "The requested resource does not exist, or you don't have permissions to access it",
	})
	return false
}

func IsAdmin(c echo.Context) bool {
	u := CurrentUser(c)
	if u == nil {
		return false
	}
	return u.IsAdmin()
}

func isAPermittedResource(permitted []string, current string) bool {
	current = strings.TrimSuffix(current, "/")
	for _, template := range permitted {
		template = strings.TrimSuffix(template, "/")
		tplParts := strings.Split(template, " ")
		curParts := strings.Split(current, " ")

		if isExactMatch(template, current) ||
			isWildcardMatch(tplParts, curParts) ||
			isAnyMatch(tplParts, curParts) {
			return true
		}
	}
	return false
}

func isExactMatch(template, current string) bool {
	return template == current
}

func isWildcardMatch(tplParts, curParts []string) bool {
	if len(tplParts) < 2 || len(curParts) < 2 {
		return false
	}

	if !strings.Contains(tplParts[1], "*") {
		return false
	}

	if curParts[0] != tplParts[0] && tplParts[0] != "ANY" {
		return false
	}

	// return strings.HasPrefix(curParts[1], strings.TrimSuffix(tplParts[1], "*"))

	s1 := "\\A" + regexp.QuoteMeta(tplParts[1]) + "\\z"
	re := regexp.MustCompile(strings.ReplaceAll(s1, "\\*", ".*"))
	return re.MatchString(curParts[1])

}

func isAnyMatch(tplParts, curParts []string) bool {
	return len(tplParts) == 2 && tplParts[0] == "ANY" &&
		len(curParts) == 2 &&
		(tplParts[1] == curParts[1] ||
			strings.HasPrefix(curParts[1], strings.TrimSuffix(tplParts[1], "*")))
}

func GenerateJWTToken(identity Identity, tokenID int, signingKey string, tokenExpiration int) (string, error) {
	// Set custom claims
	claims := &JWTCustomClaims{
		identity.GetID(),
		identity.GetName(),
		identity.IsAdmin(),
		identity.GetResources(),
		tokenID,
		identity.IsPasswordChangeRequired(),
		jwt.RegisteredClaims{
			Subject:   identity.GetID(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(tokenExpiration))),
		},
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(signingKey))
}

func GenerateRefreshToken(identity Identity, signingKey string, rTokenExpiration int) (string, error) {
	// Set custom claims
	claims := &jwt.RegisteredClaims{
		Subject:   identity.GetName(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(rTokenExpiration))),
	}

	// Create token with claims
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(signingKey))
}
