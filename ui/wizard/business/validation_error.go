// Файл validation_error.go определяет тип ошибки валидации с текстом конфига
// для копирования в буфер при ошибке sing-box check (см. saver.go).

package business

// ValidationError возвращается при ошибке валидации конфига (например sing-box check).
// ConfigText — итоговый текст конфига, который не прошёл проверку; можно скопировать для анализа.
type ValidationError struct {
	Err        error
	ConfigText string
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "validation failed"
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}
