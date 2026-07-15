package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/crypto"
	"github.com/rs/zerolog/log"
)

// DecryptMiddleware возвращает middleware, расшифровывающий тело
// запроса приватным ключом privateKey, если клиент прислал заголовок
// X-Encrypted: true. Расшифровка выполняется до gzip-middleware,
// поэтому шифротекст передавался поверх сжатия. Если privateKey nil,
// middleware пропускает запрос без изменений — это позволяет включать
// шифрование опционально на стороне агента.
func DecryptMiddleware(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil || r.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error().Err(err).Msg("ошибка чтения тела запроса при расшифровке")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			decrypted, err := crypto.Decrypt(privateKey, body)
			if err != nil {
				log.Error().Err(err).Msg("ошибка расшифровки тела запроса")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))
			next.ServeHTTP(w, r)
		})
	}
}
