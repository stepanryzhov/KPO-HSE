package apperrors

import "fmt"

// SaveError означает, что состояние в памяти могло измениться без записи на диск.
// Консоль должна завершить работу, не выполняя следующие команды.
type SaveError struct {
	Err error
}

func (e *SaveError) Error() string {
	return fmt.Sprintf("не удалось сохранить данные: %v", e.Err)
}

func (e *SaveError) Unwrap() error {
	return e.Err
}
