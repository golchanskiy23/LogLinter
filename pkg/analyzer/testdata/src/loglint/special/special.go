package special

import "log/slog"

func testSpecialChars() {
	slog.Info("user logged in")  // OK
	slog.Info("user logged in.") // want `log message must not contain special character`
	slog.Info("what happened?")  // want `log message must not contain special character`
	slog.Info("done!")           // want `log message must not contain special character`
	slog.Info("step-by-step")    // want `log message must not contain special character`
	slog.Info("🔥 fire")          // want `log message must not contain emoji`
}
