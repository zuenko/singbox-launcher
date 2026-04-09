package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"
)

const (
	// NetworkDialTimeout - таймаут на подключение к серверу
	NetworkDialTimeout = 5 * time.Second
	// NetworkRequestTimeout - таймаут на выполнение HTTP запроса
	NetworkRequestTimeout = 15 * time.Second
	// NetworkLongTimeout - таймаут для длительных операций (скачивание файлов)
	NetworkLongTimeout = 30 * time.Second
)

// CreateHTTPClient создает HTTP клиент с правильными таймаутами
// Экспортировано для использования в parsers пакете
func CreateHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   NetworkDialTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     false,
		},
	}
}

// IsNetworkError проверяет, является ли ошибка сетевой ошибкой
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Проверка на timeout
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return true
		}
		if netErr.Temporary() {
			return true
		}
	}

	// Проверка на отсутствие соединения
	if _, ok := err.(*net.OpError); ok {
		return true
	}

	// Проверка на DNS ошибку
	if _, ok := err.(*net.DNSError); ok {
		return true
	}

	// Проверка на контекст (отмена/таймаут)
	if err == context.DeadlineExceeded || err == context.Canceled {
		return true
	}

	return false
}

// GetNetworkErrorMessage возвращает понятное сообщение об ошибке сети
func GetNetworkErrorMessage(err error) string {
	if err == nil {
		return "Unknown network error"
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return "Network timeout: connection timed out"
	}

	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Op == "dial" {
			return "Network error: cannot connect to server"
		}
		return fmt.Sprintf("Network error: %s", sanitizeSensitiveURLInfo(opErr.Error()))
	}

	if dnsErr, ok := err.(*net.DNSError); ok {
		return fmt.Sprintf("DNS error: cannot resolve hostname (%s)", dnsErr.Name)
	}

	if err == context.DeadlineExceeded {
		return "Request timeout: operation took too long"
	}

	if err == context.Canceled {
		return "Request canceled"
	}

	return fmt.Sprintf("Network error: %s", sanitizeSensitiveURLInfo(err.Error()))
}

var credentialsInURLPattern = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+\-.]*://)([^:@/\s]+):([^@/\s]+)@`)

func sanitizeSensitiveURLInfo(message string) string {
	return credentialsInURLPattern.ReplaceAllString(message, "${1}${2}:***@")
}
