package english

import "log/slog"

func testEnglish() {
	slog.Info("user logged in")        // OK
	slog.Info("пользователь вошёл")    // want `log message must be in English`
	slog.Info("error: неверный токен") // want `log message must be in English`
	slog.Info("hello мир")             // want `log message must be in English`
}
