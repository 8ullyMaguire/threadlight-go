package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/model"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	pg        *pgxpool.Pool
	rdb       *redis.Client
	jwtSecret string
}

func NewAuthService(pg *pgxpool.Pool, rdb *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{pg: pg, rdb: rdb, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	existing, _ := s.findByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}
	existing, _ = s.findByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	inviteCode := generateInviteCode()

	user := &model.User{}
	err = s.pg.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, invite_code)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, display_name, bio, email, trust_level, trust_score,
		           reputation, credits, is_active, onboarding_stage, proximity_opt_out, created_at`,
		req.Username, req.Email, string(hash), inviteCode,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Email,
		&user.TrustLevel, &user.TrustScore, &user.Reputation, &user.Credits,
		&user.IsActive, &user.OnboardingStage, &user.ProximityOptOut, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.findByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return int64(userIDFloat), nil
}

func (s *AuthService) findByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, username, display_name, bio, email, password_hash, trust_level, trust_score,
		        reputation, credits, is_active, onboarding_stage, proximity_opt_out, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Email,
		&user.PasswordHash, &user.TrustLevel, &user.TrustScore, &user.Reputation,
		&user.Credits, &user.IsActive, &user.OnboardingStage, &user.ProximityOptOut, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) findByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	err := s.pg.QueryRow(ctx,
		`SELECT id, username, display_name, bio, email, password_hash, trust_level, trust_score,
		        reputation, credits, is_active, onboarding_stage, proximity_opt_out, created_at
		 FROM users WHERE username = $1`, username,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Email,
		&user.PasswordHash, &user.TrustLevel, &user.TrustScore, &user.Reputation,
		&user.Credits, &user.IsActive, &user.OnboardingStage, &user.ProximityOptOut, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateInviteCode() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
