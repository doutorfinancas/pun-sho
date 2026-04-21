package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/doutorfinancas/pun-sho/entity"
)

type AuthService struct {
	log             *zap.Logger
	userRepo        *entity.UserRepository
	sessionRepo     *entity.SessionRepository
	sessionDuration time.Duration
}

func NewAuthService(
	log *zap.Logger,
	userRepo *entity.UserRepository,
	sessionRepo *entity.SessionRepository,
	sessionDuration time.Duration,
) *AuthService {
	if sessionDuration == 0 {
		sessionDuration = 48 * time.Hour
	}

	return &AuthService{
		log:             log,
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		sessionDuration: sessionDuration,
	}
}

func (s *AuthService) Login(username, password string) (*entity.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *AuthService) ValidateTOTP(user *entity.User, code string) bool {
	if !user.TOTPEnabled || user.TOTPSecret == "" {
		return true
	}

	return totp.Validate(code, user.TOTPSecret)
}

func (s *AuthService) RegisterTOTP(user *entity.User) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "pun-sho",
		AccountName: user.Username,
	})
	if err != nil {
		return "", "", err
	}

	user.TOTPSecret = key.Secret()
	user.TOTPEnabled = true

	if err := s.userRepo.Database.Save(user); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *AuthService) CreateSession(userID uuid.UUID) (*entity.Session, error) {
	return s.createSession(userID, true)
}

func (s *AuthService) CreatePendingSession(userID uuid.UUID) (*entity.Session, error) {
	return s.createSession(userID, false)
}

func (s *AuthService) createSession(userID uuid.UUID, verified bool) (*entity.Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}

	session := &entity.Session{
		UserID:    userID,
		Token:     hex.EncodeToString(tokenBytes),
		Verified:  verified,
		ExpiresAt: time.Now().Add(s.sessionDuration),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) ValidatePendingSession(token string) (*entity.User, error) {
	session, err := s.sessionRepo.FindByToken(token)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *AuthService) VerifySession(token string) error {
	session, err := s.sessionRepo.FindByToken(token)
	if err != nil {
		return errors.New("invalid session")
	}

	session.Verified = true
	return s.sessionRepo.Database.Save(session)
}

func (s *AuthService) ValidateSession(token string) (*entity.User, error) {
	session, err := s.sessionRepo.FindByToken(token)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	if !session.Verified {
		return nil, errors.New("session not verified")
	}

	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *AuthService) Logout(token string) error {
	return s.sessionRepo.DeleteByToken(token)
}

func (s *AuthService) CreateUser(username, email, password, role string) (*entity.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ResetPassword(userID uuid.UUID, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Database.Save(user)
}

func (s *AuthService) ToggleRole(userID uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Role == "admin" {
		user.Role = "user"
	} else {
		user.Role = "admin"
	}

	if err := s.userRepo.Database.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) DeleteUser(userID uuid.UUID) error {
	// Hard delete — also remove associated sessions
	if err := s.sessionRepo.DeleteByUserID(userID); err != nil {
		s.log.Warn("Failed to delete user sessions", zap.Error(err))
	}

	return s.userRepo.HardDelete(userID)
}

func (s *AuthService) ListUsers() ([]*entity.User, error) {
	return s.userRepo.List()
}

func (s *AuthService) UserCount() (int64, error) {
	return s.userRepo.Count()
}

func (s *AuthService) FindOrCreateMSUser(email, msEmail string) (*entity.User, error) {
	user, err := s.userRepo.FindByMSEmail(msEmail)
	if err == nil {
		return user, nil
	}

	// Try finding by email
	user, err = s.userRepo.FindByEmail(email)
	if err == nil {
		user.MSLinked = true
		user.MSEmail = &msEmail
		if saveErr := s.userRepo.Database.Save(user); saveErr != nil {
			return nil, saveErr
		}
		return user, nil
	}

	// Create new user from MS account
	user = &entity.User{
		Username:     email,
		Email:        email,
		PasswordHash: "",
		MSLinked:     true,
		MSEmail:      &msEmail,
		Role:         "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) SeedDefaultAdmin(password string) error {
	count, err := s.userRepo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	s.log.Warn("Creating default admin user - please change the password after first login")

	_, err = s.CreateUser("admin", "admin@pun-sho.local", password, "admin")
	return err
}
