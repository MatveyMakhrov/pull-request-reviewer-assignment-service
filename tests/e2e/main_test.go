// tests/e2e/main_test.go
package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	baseURL = "http://localhost:8082"
)

type E2ETestSuite struct {
	suite.Suite
	client *http.Client
}

func (suite *E2ETestSuite) SetupSuite() {
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Очищаем БД перед всеми тестами
	err := CleanTestDatabase()
	if err != nil {
		suite.T().Logf("Warning: failed to clean database: %v", err)
	}

	// Ждем пока сервис поднимется
	suite.waitForService()
}

func (suite *E2ETestSuite) TearDownTest() {
	// Очищаем БД после каждого теста
	err := CleanTestDatabase()
	if err != nil {
		suite.T().Logf("Warning: failed to clean database: %v", err)
	}
}

func (suite *E2ETestSuite) waitForService() {
	for i := 0; i < 30; i++ {
		resp, err := suite.client.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	suite.T().Fatal("Service didn't start in time")
}

func (suite *E2ETestSuite) Test_CompleteWorkflow() {
	t := suite.T()

	// === 1. Создание команды ===
	team := map[string]interface{}{
		"team_name": "e2e-development-team",
		"members": []map[string]interface{}{
			{
				"user_id":   "dev-lead",
				"username":  "Alice Developer",
				"is_active": true,
			},
			{
				"user_id":   "senior-dev",
				"username":  "Bob Senior",
				"is_active": true,
			},
			{
				"user_id":   "mid-dev",
				"username":  "Charlie Middle",
				"is_active": true,
			},
			{
				"user_id":   "junior-dev",
				"username":  "David Junior",
				"is_active": true,
			},
			{
				"user_id":   "inactive-dev",
				"username":  "Eve Inactive",
				"is_active": false,
			},
		},
	}

	statusCode, body, err := suite.makeRequest("POST", "/team/add", team)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Создание команды должно быть успешным")

	var teamResponse map[string]interface{}
	err = json.Unmarshal(body, &teamResponse)
	assert.NoError(t, err)
	assert.Contains(t, teamResponse, "team")

	// === 2. Создание первого PR ===
	pr1 := map[string]string{
		"pull_request_id":   "e2e-feature-auth",
		"pull_request_name": "Implement authentication system",
		"author_id":         "dev-lead",
	}

	statusCode, body, err = suite.makeRequest("POST", "/pullRequest/create", pr1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, statusCode, "Создание PR должно быть успешным")

	var pr1Response map[string]interface{}
	err = json.Unmarshal(body, &pr1Response)
	assert.NoError(t, err)
	assert.Contains(t, pr1Response, "pr")

	// Проверяем что назначились ревьюверы
	pr1Data := pr1Response["pr"].(map[string]interface{})
	reviewers := pr1Data["assigned_reviewers"].([]interface{})
	assert.Greater(t, len(reviewers), 0, "Должны быть назначены ревьюверы")
	assert.LessOrEqual(t, len(reviewers), 2, "Не более 2 ревьюверов")

	// === 3. Создание второго PR ===
	pr2 := map[string]string{
		"pull_request_id":   "e2e-feature-api",
		"pull_request_name": "Add REST API endpoints",
		"author_id":         "senior-dev",
	}

	statusCode, _, err = suite.makeRequest("POST", "/pullRequest/create", pr2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, statusCode)

	// === 4. Получение PR для ревью пользователем ===
	statusCode, body, err = suite.makeGetRequest("/users/getReview?user_id=mid-dev")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)

	var userPRsResponse map[string]interface{}
	err = json.Unmarshal(body, &userPRsResponse)
	assert.NoError(t, err)
	assert.Contains(t, userPRsResponse, "pull_requests")

	// === 5. Переназначение ревьювера ===
	reassignRequest := map[string]string{
		"pull_request_id": "e2e-feature-auth",
		"old_user_id":     reviewers[0].(string),
	}

	statusCode, body, err = suite.makeRequest("POST", "/pullRequest/reassign", reassignRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Переназначение должно быть успешным")

	var reassignResponse map[string]interface{}
	err = json.Unmarshal(body, &reassignResponse)
	assert.NoError(t, err)
	assert.Contains(t, reassignResponse, "replaced_by")

	// === 6. Мерж PR ===
	mergeRequest := map[string]string{
		"pull_request_id": "e2e-feature-auth",
	}

	statusCode, _, err = suite.makeRequest("POST", "/pullRequest/merge", mergeRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Мерж PR должен быть успешным")

	// === 7. Проверка идемпотентности мержа ===
	statusCode, _, err = suite.makeRequest("POST", "/pullRequest/merge", mergeRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Повторный мерж должен быть успешным (идемпотентность)")

	// === 8. Попытка переназначения в замерженном PR (должна быть ошибка) ===
	statusCode, _, err = suite.makeRequest("POST", "/pullRequest/reassign", reassignRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, statusCode, "Нельзя переназначать в замерженном PR")

	// === 9. Массовая деактивация пользователей ===
	bulkDeactivateRequest := map[string]interface{}{
		"team_name": "e2e-development-team",
		"user_ids":  []string{"mid-dev", "junior-dev"},
	}

	statusCode, body, err = suite.makeRequest("POST", "/users/bulk-deactivate", bulkDeactivateRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Массовая деактивация должна быть успешной")

	var bulkResponse map[string]interface{}
	err = json.Unmarshal(body, &bulkResponse)
	assert.NoError(t, err)
	assert.Contains(t, bulkResponse, "deactivated_users")

	// === 10. Создание PR после деактивации ===
	pr3 := map[string]string{
		"pull_request_id":   "e2e-feature-ui",
		"pull_request_name": "Add user interface",
		"author_id":         "dev-lead",
	}

	statusCode, _, err = suite.makeRequest("POST", "/pullRequest/create", pr3)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, statusCode)

	// === 11. Получение статистики ===
	statusCode, body, err = suite.makeGetRequest("/stats/review-assignments")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode, "Статистика должна быть доступна")

	var statsResponse map[string]interface{}
	err = json.Unmarshal(body, &statsResponse)
	assert.NoError(t, err)

	assert.Contains(t, statsResponse, "total_assignments")
	assert.Contains(t, statsResponse, "assignments_by_user")
	assert.Contains(t, statsResponse, "assignments_by_pr")
	assert.Contains(t, statsResponse, "top_reviewers")

	// === 12. Получение информации о команде ===
	statusCode, body, err = suite.makeGetRequest("/team/get?team_name=e2e-development-team")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)

	var teamInfoResponse map[string]interface{}
	err = json.Unmarshal(body, &teamInfoResponse)
	assert.NoError(t, err)
	assert.Contains(t, teamInfoResponse, "team")
}

func (suite *E2ETestSuite) Test_ErrorScenarios() {
	t := suite.T()

	// === 1. Создание PR с несуществующим автором ===
	invalidPR := map[string]string{
		"pull_request_id":   "e2e-invalid-1",
		"pull_request_name": "Invalid PR",
		"author_id":         "non-existent-user-12345",
	}

	statusCode, _, err := suite.makeRequest("POST", "/pullRequest/create", invalidPR)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, statusCode, "Должна быть ошибка для несуществующего автора")

	// === 2. Создание команды с пустым названием ===
	invalidTeam := map[string]interface{}{
		"team_name": "",
		"members": []map[string]interface{}{
			{
				"user_id":   "test-user-12345",
				"username":  "Test User",
				"is_active": true,
			},
		},
	}

	statusCode, _, err = suite.makeRequest("POST", "/team/add", invalidTeam)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, statusCode, "Должна быть ошибка валидации")

	// === 3. Неверный HTTP метод ===
	req, err := http.NewRequest("PUT", baseURL+"/team/add", nil)
	assert.NoError(t, err)
	resp, err := suite.client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "Должна быть ошибка метода")
}

// Вспомогательные методы остаются без изменений
func (suite *E2ETestSuite) makeRequest(method, path string, body interface{}) (int, []byte, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, respBody, nil
}

func (suite *E2ETestSuite) makeGetRequest(path string) (int, []byte, error) {
	resp, err := suite.client.Get(baseURL + path)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, body, nil
}

func TestE2ESuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	suite.Run(t, new(E2ETestSuite))
}
