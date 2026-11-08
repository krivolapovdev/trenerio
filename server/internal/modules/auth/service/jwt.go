package service

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/krivolapovdev/trenerio/internal/modules/auth/domain"
)

type JwtUser = domain.JwtUser
type Tokens = domain.Tokens
type LoginResult = domain.LoginResult

const (
	claimSubject   = "sub"
	claimEmail     = "email"
	claimType      = "type"
	claimIssuedAt  = "iat"
	claimExpiresAt = "exp"
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidTokenClaims      = errors.New("invalid token claims")
	ErrMissingSubjectClaim     = errors.New("missing sub claim")
	ErrInvalidSubjectFormat    = errors.New("invalid sub format")
)

type TokenProvider struct {
	secretKey              []byte
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
	reviewerKey            string
}

func NewTokenProvider(secretKey string, accessExp, refreshExp time.Duration, reviewerKey string) *TokenProvider {
	if accessExp == 0 {
		accessExp = 15 * time.Minute
	}

	if refreshExp == 0 {
		refreshExp = 14 * 24 * time.Hour
	}

	return &TokenProvider{
		secretKey:              []byte(secretKey),
		accessTokenExpiration:  accessExp,
		refreshTokenExpiration: refreshExp,
		reviewerKey:            reviewerKey,
	}
}

func (p *TokenProvider) ValidateReviewerKey(key string) bool {
	if p.reviewerKey == "" || key == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(p.reviewerKey)) == 1
}

func (p *TokenProvider) GenerateAccessToken(user JwtUser) (string, error) {
	return p.generateToken(user, p.accessTokenExpiration, TokenTypeAccess)
}

func (p *TokenProvider) GenerateRefreshToken(user JwtUser) (string, error) {
	return p.generateToken(user, p.refreshTokenExpiration, TokenTypeRefresh)
}

func (p *TokenProvider) GenerateTokenPair(user JwtUser) (accessToken, refreshToken string, err error) {
	accessToken, err = p.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = p.GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (p *TokenProvider) generateToken(user JwtUser, exp time.Duration, tokenType TokenType) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		claimSubject:   user.ID.String(),
		claimEmail:     user.Email,
		claimType:      string(tokenType),
		claimIssuedAt:  now.Unix(),
		claimExpiresAt: now.Add(exp).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.secretKey)
}

func (p *TokenProvider) IsTokenValid(tokenStr string, expectedType TokenType) bool {
	token, err := p.parseToken(tokenStr)
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	tokenType, ok := claims[claimType].(string)
	if !ok || tokenType != string(expectedType) {
		return false
	}

	return true
}

func (p *TokenProvider) ExtractUser(tokenStr string) (JwtUser, error) {
	token, err := p.parseToken(tokenStr)
	if err != nil || !token.Valid {
		return JwtUser{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return JwtUser{}, ErrInvalidTokenClaims
	}

	subStr, ok := claims[claimSubject].(string)
	if !ok {
		return JwtUser{}, ErrMissingSubjectClaim
	}

	userID, err := uuid.Parse(subStr)
	if err != nil {
		return JwtUser{}, ErrInvalidSubjectFormat
	}

	email, _ := claims[claimEmail].(string)

	return JwtUser{
		ID:    userID,
		Email: email,
	}, nil
}

func (p *TokenProvider) parseToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return p.secretKey, nil
	})
}
