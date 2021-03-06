package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/artifacthub/hub/internal/hub"
	"github.com/artifacthub/hub/internal/tests"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var (
	errFakeDatabaseFailure    = errors.New("fake database failure")
	errFakeEmailSenderFailure = errors.New("fake email sender failure")
)

func TestCheckCredentials(t *testing.T) {
	dbQuery := `select user_id, password from "user" where email = $1`

	t.Run("credentials provided not found in database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return(nil, pgx.ErrNoRows)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("error getting credentials from database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return(nil, errFakeDatabaseFailure)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.Equal(t, errFakeDatabaseFailure, err)
		assert.Nil(t, output)
		db.AssertExpectations(t)
	})

	t.Run("invalid credentials provided", func(t *testing.T) {
		pw, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return([]interface{}{"userID", string(pw)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass2")
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("valid credentials provided", func(t *testing.T) {
		pw, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "email").Return([]interface{}{"userID", string(pw)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckCredentials(context.Background(), "email", "pass")
		assert.NoError(t, err)
		assert.True(t, output.Valid)
		assert.Equal(t, "userID", output.UserID)
		db.AssertExpectations(t)
	})
}

func TestCheckSession(t *testing.T) {
	dbQuery := `
	select user_id, floor(extract(epoch from created_at))
	from session where session_id = $1
	`

	t.Run("session not found in database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return(nil, pgx.ErrNoRows)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("error getting session from database", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return(nil, errFakeDatabaseFailure)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.Equal(t, errFakeDatabaseFailure, err)
		assert.Nil(t, output)
		db.AssertExpectations(t)
	})

	t.Run("session has expired", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return([]interface{}{"userID", int64(1)}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, output.Valid)
		assert.Empty(t, output.UserID)
		db.AssertExpectations(t)
	})

	t.Run("valid session", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, []byte("sessionID")).Return([]interface{}{
			"userID",
			time.Now().Unix(),
		}, nil)
		m := NewManager(db, nil)

		output, err := m.CheckSession(context.Background(), []byte("sessionID"), 1*time.Hour)
		assert.NoError(t, err)
		assert.True(t, output.Valid)
		assert.Equal(t, "userID", output.UserID)
		db.AssertExpectations(t)
	})
}

func TestDeleteSession(t *testing.T) {
	dbQuery := "delete from session where session_id = $1"

	t.Run("delete session", func(t *testing.T) {
		testCases := []struct {
			description string
			dbResponse  interface{}
		}{
			{
				"session deleted successfully",
				nil,
			},
			{
				"error deleting session from database",
				errFakeDatabaseFailure,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				db := &tests.DBMock{}
				db.On("Exec", dbQuery, []byte("sessionID")).Return(tc.dbResponse)
				m := NewManager(db, nil)

				err := m.DeleteSession(context.Background(), []byte("sessionID"))
				assert.Equal(t, tc.dbResponse, err)
				db.AssertExpectations(t)
			})
		}
	})
}

func TestGetAlias(t *testing.T) {
	dbQuery := `select alias from "user" where user_id = $1`
	ctx := context.WithValue(context.Background(), hub.UserIDKey, "userID")

	t.Run("database query succeeded", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "userID").Return("alias", nil)
		m := NewManager(db, nil)

		alias, err := m.GetAlias(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "alias", alias)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "userID").Return("", errFakeDatabaseFailure)
		m := NewManager(db, nil)

		alias, err := m.GetAlias(ctx)
		assert.Equal(t, errFakeDatabaseFailure, err)
		assert.Empty(t, alias)
		db.AssertExpectations(t)
	})
}

func TestRegisterSession(t *testing.T) {
	dbQuery := "select register_session($1::jsonb)"

	s := &hub.Session{
		UserID:    "00000000-0000-0000-0000-000000000001",
		IP:        "192.168.1.100",
		UserAgent: "Safari 13.0.5",
	}

	t.Run("successful session registration", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return([]byte("sessionID"), nil)
		m := NewManager(db, nil)

		sessionID, err := m.RegisterSession(context.Background(), s)
		assert.NoError(t, err)
		assert.Equal(t, []byte("sessionID"), sessionID)
		db.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return(nil, errFakeDatabaseFailure)
		m := NewManager(db, nil)

		sessionID, err := m.RegisterSession(context.Background(), s)
		assert.Equal(t, errFakeDatabaseFailure, err)
		assert.Nil(t, sessionID)
		db.AssertExpectations(t)
	})
}

func TestRegisterUser(t *testing.T) {
	dbQuery := "select register_user($1::jsonb)"

	t.Run("successful user registration in database", func(t *testing.T) {
		testCases := []struct {
			description         string
			emailSenderResponse error
		}{
			{
				"email verification code sent successfully",
				nil,
			},
			{
				"error sending email verification code",
				errFakeEmailSenderFailure,
			},
		}
		for _, tc := range testCases {
			tc := tc
			t.Run(tc.description, func(t *testing.T) {
				db := &tests.DBMock{}
				db.On("QueryRow", dbQuery, mock.Anything).Return("emailVerificationCode", nil)
				es := &tests.EmailSenderMock{}
				es.On("SendEmail", mock.Anything).Return(tc.emailSenderResponse)
				m := NewManager(db, es)

				u := &hub.User{
					Alias:     "alias",
					FirstName: "first_name",
					LastName:  "last_name",
					Email:     "email@email.com",
					Password:  "password",
				}
				err := m.RegisterUser(context.Background(), u, "")
				assert.Equal(t, tc.emailSenderResponse, err)
				assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("password")))
				db.AssertExpectations(t)
				es.AssertExpectations(t)
			})
		}
	})

	t.Run("database error registering user", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, mock.Anything).Return("", errFakeDatabaseFailure)
		m := NewManager(db, nil)

		err := m.RegisterUser(context.Background(), &hub.User{}, "")
		assert.Equal(t, errFakeDatabaseFailure, err)
		db.AssertExpectations(t)
	})
}

func TestVerifyEmail(t *testing.T) {
	dbQuery := "select verify_email($1::uuid)"

	t.Run("successful email verification", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "emailVerificationCode").Return(true, nil)
		m := NewManager(db, nil)

		verified, err := m.VerifyEmail(context.Background(), "emailVerificationCode")
		assert.NoError(t, err)
		assert.True(t, verified)
		db.AssertExpectations(t)
	})

	t.Run("database error verifying email", func(t *testing.T) {
		db := &tests.DBMock{}
		db.On("QueryRow", dbQuery, "emailVerificationCode").Return(false, errFakeDatabaseFailure)
		m := NewManager(db, nil)

		verified, err := m.VerifyEmail(context.Background(), "emailVerificationCode")
		assert.Equal(t, errFakeDatabaseFailure, err)
		assert.False(t, verified)
		db.AssertExpectations(t)
	})
}
