package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityLogger логирует подозрительные запросы
func SecurityLogger(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")
		userAgent := c.GetHeader("User-Agent")

		// Проверяем, является ли origin доверенным
		isTrusted := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				isTrusted = true
				break
			}
		}

		// Логируем все запросы с неизвестных доменов
		if !isTrusted && origin != "" {
			log.Printf("⚠️  SECURITY: Unknown origin: %s, Referer: %s, UA: %s, IP: %s, Path: %s",
				origin, referer, userAgent, c.ClientIP(), c.Request.URL.Path)
		}

		c.Next()
	}
}

// RefererCheck проверяет Referer заголовок для дополнительной защиты
func RefererCheck(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем OPTIONS запросы
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Пропускаем health check и публичные эндпоинты
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/api/public") {
			c.Next()
			return
		}

		referer := c.GetHeader("Referer")
		origin := c.GetHeader("Origin")

		// Если нет ни referer, ни origin - подозрительно, но можем пропустить для прямых API запросов
		if referer == "" && origin == "" {
			// Логируем для мониторинга
			log.Printf("⚠️  SECURITY: No Referer/Origin from IP: %s, Path: %s, UA: %s",
				c.ClientIP(), c.Request.URL.Path, c.GetHeader("User-Agent"))
			c.Next()
			return
		}

		// Проверяем, что referer или origin из доверенного домена
		isValid := false
		for _, domain := range allowedOrigins {
			if strings.HasPrefix(referer, domain) || origin == domain {
				isValid = true
				break
			}
		}

		if !isValid && (referer != "" || origin != "") {
			log.Printf("🚨 SECURITY ALERT: Blocked suspicious referer/origin - Referer: %s, Origin: %s, IP: %s, Path: %s",
				referer, origin, c.ClientIP(), c.Request.URL.Path)
			c.JSON(403, gin.H{
				"error": "forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HTTPSEnforcement перенаправляет HTTP на HTTPS в production
func HTTPSEnforcement(environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if environment == "production" {
			// Проверяем X-Forwarded-Proto заголовок (используется балансировщиками)
			proto := c.GetHeader("X-Forwarded-Proto")
			if proto != "" && proto != "https" {
				httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
				c.Redirect(http.StatusMovedPermanently, httpsURL)
				c.Abort()
				return
			}

			// Добавляем HSTS заголовок
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}

// SecurityHeaders добавляет заголовки безопасности
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Предотвращает MIME type sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// Защита от clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// XSS защита (для старых браузеров)
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Контроль Referer
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - базовая политика
		// Настройте под ваши нужды
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Permissions Policy (бывший Feature-Policy)
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
