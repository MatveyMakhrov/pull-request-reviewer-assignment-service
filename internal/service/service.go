package service

import "fmt"

// представляет структуру для стандартизированных ошибок логики
type ServiceError struct {
	Code    string
	Message string
}

// возвращает строковое представление ошибки сервиса в формате "КОД: сообщение"
// принимает: не принемает параметров, работает с получателем ServiceError
// возвращает: строку с отформатированным сообщением об ошибке
func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// создает новый экземпляр стандартизированной ошибки логики
// принимает: код ошибки и текстовое сообщение для инициализации
// возвращает: указатель на созданный объект ServiceError
func NewServiceError(code, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}
