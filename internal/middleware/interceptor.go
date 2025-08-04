package middleware

import (
	"net"
	"net/http"
)

// TrustedSubnet проверяет, что запрос пришел из доверенной подсети.
func TrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем доверенную подсеть
			if trustedSubnet == "" {
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			// Получаем IP из заголовка X-Real-IP
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			// Парсим IP
			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			// Парсим доверенную подсеть
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			// Проверяем принадлежность IP к подсети
			if !subnet.Contains(ip) {
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
