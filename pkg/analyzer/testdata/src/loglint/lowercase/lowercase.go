package lowercase

import "log/slog"

func testLowercase() {
	slog.Info("everything is fine")  // OK
	slog.Info("Starting server")     // want `log message must start with lowercase`
	slog.Error("Failed to connect")  // want `log message must start with lowercase`
	slog.Info("")                    // OK — пустая строка, checkLowercase пропускает
	slog.Info("123 items processed") // OK — первый символ не буква
}
