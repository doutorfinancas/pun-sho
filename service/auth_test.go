package service

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/doutorfinancas/pun-sho/database"
	"github.com/doutorfinancas/pun-sho/entity"
	mockDB "github.com/doutorfinancas/pun-sho/test"
)

func setupAuthTest() (*AuthService, sqlmock.Sqlmock) {
	mock, gormDB := mockDB.NewMockDB()
	log, _ := zap.NewDevelopment()

	db := database.NewDatabase(gormDB)
	userRepo := entity.NewUserRepository(db, log)
	sessionRepo := entity.NewSessionRepository(db, log)

	authSvc := NewAuthService(log, userRepo, sessionRepo, 48*time.Hour)

	return authSvc, mock
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSvc, mock := setupAuthTest()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 AND deleted_at IS NULL`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{}))

	_, err := authSvc.Login("nonexistent", "password")
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
}

func TestAuthService_ValidateSession_Invalid(t *testing.T) {
	authSvc, mock := setupAuthTest()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sessions" WHERE token = $1 AND expires_at > $2`)).
		WithArgs("invalid_token", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{}))

	_, err := authSvc.ValidateSession("invalid_token")
	assert.Error(t, err)
	assert.Equal(t, "invalid session", err.Error())
}

func TestAuthService_DefaultSessionDuration(t *testing.T) {
	log, _ := zap.NewDevelopment()
	_, gormDB := mockDB.NewMockDB()

	db := database.NewDatabase(gormDB)
	userRepo := entity.NewUserRepository(db, log)
	sessionRepo := entity.NewSessionRepository(db, log)

	authSvc := NewAuthService(log, userRepo, sessionRepo, 0)
	assert.Equal(t, 48*time.Hour, authSvc.sessionDuration)
}

func TestAuthService_SeedDefaultAdmin_UsersExist(t *testing.T) {
	authSvc, mock := setupAuthTest()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users" WHERE deleted_at IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	err := authSvc.SeedDefaultAdmin("admin_password")
	assert.NoError(t, err)
}

func TestAuthService_ValidateTOTP_Disabled(t *testing.T) {
	authSvc, _ := setupAuthTest()

	user := &entity.User{
		ID:          uuid.New(),
		TOTPEnabled: false,
	}

	assert.True(t, authSvc.ValidateTOTP(user, ""))
	assert.True(t, authSvc.ValidateTOTP(user, "123456"))
}

func TestAuthService_ValidateTOTP_InvalidCode(t *testing.T) {
	authSvc, _ := setupAuthTest()

	user := &entity.User{
		ID:          uuid.New(),
		TOTPEnabled: true,
		TOTPSecret:  "JBSWY3DPEHPK3PXP",
	}

	assert.False(t, authSvc.ValidateTOTP(user, "000000"))
}
