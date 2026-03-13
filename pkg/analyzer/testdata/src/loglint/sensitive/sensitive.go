package sensitive

import "log/slog"

func testSensitive() {
	slog.Info("user authenticated")        // OK
	slog.Info("user logged in")            // OK
	slog.Info("invalid password provided") // want `log message may contain sensitive data`
	slog.Info("loading token from env")    // want `log message may contain sensitive data`
	slog.Info("auth failed")               // want `log message may contain sensitive data`

	password := "secret123"
	slog.Info("login attempt: " + password) // want `log message concatenates potentially sensitive`
}
