package controllers

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"subscriptions/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	argsList := m.Called(append([]interface{}{query}, args...)...)
	if result, ok := argsList.Get(0).(*sql.Rows); ok {
		return result, argsList.Error(1)
	}
	return nil, argsList.Error(1)
}

func TestGetSubscriptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing sqlmock: %v", err)
	}
	defer db.Close()

	testCases := []struct {
		name         string
		mockData     [][]interface{}
		mockError    error
		expectedLen  int
		expectedCode int
	}{
		{
			name: "success - subscriptions found",
			mockData: [][]interface{}{
				{1, "Sub1", 101, 5, time.Now(), time.Now(), nil},
				{2, "Sub2", 102, 10, time.Now(), time.Now(), nil},
			},
			expectedLen:  2,
			expectedCode: http.StatusOK,
		},
		{
			name:         "no subscriptions found",
			mockData:     [][]interface{}{},
			expectedLen:  0,
			expectedCode: http.StatusOK,
		},
		{
			name:         "database error",
			mockData:     nil,
			mockError:    errors.New("database error"),
			expectedLen:  0,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := `SELECT \* FROM subscriptions WHERE deleted_at IS NULL`

			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "name", "product_id", "license_count", "created_at", "updated_at", "deleted_at"})
				for _, row := range tc.mockData {
					var values []driver.Value
					for _, v := range row {
						values = append(values, v)
					}
					rows.AddRow(values...)
				}
				mock.ExpectQuery(query).WillReturnRows(rows)
			}

			req := httptest.NewRequest("GET", "/subscriptions", nil)
			w := httptest.NewRecorder()

			handler := GetSubscriptions(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.mockError == nil {
				var subscriptions []models.Subscription
				if err := json.NewDecoder(w.Body).Decode(&subscriptions); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				assert.Len(t, subscriptions, tc.expectedLen)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSubscriptionByID(t *testing.T) {
	type testCase struct {
		name      string
		subID     string
		mockData  []interface{}
		expectErr bool
		mockError error
	}

	testCases := []testCase{
		{
			name:  "success - valid subscription",
			subID: "1",
			mockData: []interface{}{
				1, "Basic Plan", 101, 10, time.Now(), time.Now(), nil,
			},
			expectErr: false,
		},
		{
			name:      "subscription not found",
			subID:     "99",
			mockData:  nil,
			expectErr: true,
		},
		{
			name:      "database error",
			subID:     "1",
			mockData:  nil,
			mockError: errors.New("database error"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)

			query := regexp.QuoteMeta(`SELECT * FROM subscriptions WHERE id = $1 AND deleted_at IS NULL`)

			if tc.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tc.subID).WillReturnError(tc.mockError)
			} else if tc.mockData != nil {
				rowValues := make([]driver.Value, len(tc.mockData))
				for i, v := range tc.mockData {
					rowValues[i] = v
				}

				rows := sqlmock.NewRows([]string{"id", "name", "product_id", "license_count", "created_at", "updated_at", "deleted_at"}).
					AddRow(rowValues...)

				mock.ExpectQuery(query).WithArgs(tc.subID).WillReturnRows(rows).RowsWillBeClosed()
			}

			req := httptest.NewRequest("GET", "/subscription/"+tc.subID, nil)
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.subID})

			handler := GetSubscriptionByID(db)
			handler.ServeHTTP(w, req)

			// Debugging logs
			fmt.Println("Response Code:", w.Code)
			fmt.Println("Response Body:", w.Body.String())

			if tc.expectErr {
				assert.Equal(t, http.StatusNotFound, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)

				var subscription models.Subscription
				err := json.NewDecoder(w.Body).Decode(&subscription)
				assert.NoError(t, err)
				assert.Equal(t, tc.mockData[1], subscription.Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
		mockQueries  func()
	}{
		{
			name:         "success - valid request",
			requestBody:  `{"name": "Premium Subscription", "product_id": 101, "license_count": 10}`,
			expectedCode: http.StatusCreated,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO subscriptions \(name, product_id, license_count\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
					WithArgs("Premium Subscription", 101, 10).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(1, time.Now(), time.Now()))
			},
		},
		{
			name:         "failure - invalid JSON",
			requestBody:  `{"name": }`, // Malformed JSON
			expectedCode: http.StatusBadRequest,
			mockQueries: func() {
				// No DB queries should run because JSON is invalid
			},
		},
		{
			name:         "failure - database error on insert",
			requestBody:  `{"name": "Standard Subscription", "product_id": 102, "license_count": 5}`,
			expectedCode: http.StatusInternalServerError,
			mockQueries: func() {
				mock.ExpectQuery(`INSERT INTO subscriptions \(name, product_id, license_count\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at`).
					WithArgs("Standard Subscription", 102, 5).
					WillReturnError(errors.New("insert error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockQueries()

			req := httptest.NewRequest("POST", "/subscriptions", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := CreateSubscription(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		subscriptionID string
		requestBody  string
		expectedCode int
		mockQueries  func()
	}{
		{
			name:         "success - valid request",
			subscriptionID: "1",
			requestBody:  `{"name": "Updated Subscription Name", "product_id": 2, "license_count": 5}`,
			expectedCode: http.StatusOK,
			mockQueries: func() {
				mock.ExpectExec(`UPDATE subscriptions SET name = \$1, product_id = \$2, license_count = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4 AND deleted_at IS NULL`).
					WithArgs("Updated Subscription Name", 2, 5, "1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:         "failure - invalid JSON",
			subscriptionID: "1",
			requestBody:  `{"name": }`, // Malformed JSON
			expectedCode: http.StatusBadRequest,
			mockQueries:  func() {}, // No DB queries should run because JSON is invalid
		},
		{
			name:         "failure - database error on update",
			subscriptionID: "1",
			requestBody:  `{"name": "New Subscription Name", "product_id": 3, "license_count": 10}`,
			expectedCode: http.StatusInternalServerError,
			mockQueries: func() {
				mock.ExpectExec(`UPDATE subscriptions SET name = \$1, product_id = \$2, license_count = \$3, updated_at = CURRENT_TIMESTAMP WHERE id = \$4 AND deleted_at IS NULL`).
					WithArgs("New Subscription Name", 3, 10, "1").
					WillReturnError(errors.New("update error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockQueries()

			req := httptest.NewRequest("PUT", "/subscriptions/"+tc.subscriptionID, strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			req = mux.SetURLVars(req, map[string]string{"id": tc.subscriptionID})

			handler := UpdateSubscription(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	testCases := []struct {
		name         string
		subscriptionID string // Now as a string since it's passed from the URL
		expectedCode int
		mockExec     func()
	}{
		{
			name:         "success - subscription deleted",
			subscriptionID: "1",
			expectedCode: http.StatusNoContent,
			mockExec: func() {
				mock.ExpectExec(`UPDATE subscriptions SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
			},
		},
		{
			name:         "failure - database error",
			subscriptionID: "1",
			expectedCode: http.StatusInternalServerError,
			mockExec: func() {
				mock.ExpectExec(`UPDATE subscriptions SET deleted_at = CURRENT_TIMESTAMP WHERE id = \$1`).
					WithArgs("1").
					WillReturnError(errors.New("database error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockExec()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/subscription/%s", tc.subscriptionID), nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.subscriptionID})
			w := httptest.NewRecorder()

			handler := DeleteSubscription(db)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)

			err := mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}







