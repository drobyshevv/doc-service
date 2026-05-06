package handler

import "net/http"

// Health обрабатывает HTTP-запрос
// проверки доступности сервиса.
//
// Используется для healthcheck и мониторинга.
func Health(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
