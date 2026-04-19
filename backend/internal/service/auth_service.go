package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"auction-platform/internal/config"
	"auction-platform/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidToken     = errors.New("invalid token")
)

type AuthService struct {
	userRepo UserRepo
	jwtCfg   *config.JWTConfig
}

func NewAuthService(userRepo UserRepo, jwtCfg *config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(username, password, email, role string) (*model.User, error) {
	exists, err := s.userRepo.ExistsByUsername(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	exists, err = s.userRepo.ExistsByEmail(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Role:     role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtCfg.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
