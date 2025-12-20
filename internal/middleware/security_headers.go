package middleware

import "github.com/gofiber/fiber/v2"

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection (legacy browsers)
		c.Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature Policy)
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		// Note: Adjust CSP based on your needs
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // React dev mode needs unsafe-inline/eval
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss:; " + // Allow WebSocket
			"frame-ancestors 'none'"
		c.Set("Content-Security-Policy", csp)

		return c.Next()
	}
}
