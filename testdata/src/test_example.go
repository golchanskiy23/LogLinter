package main

import (
	"log/slog"

	"go.uber.org/zap"
)

func main() {
	// Примеры логов для тестирования
	slog.Info("User login successful")     // OK
	slog.Info("User login failed")         // OK
	slog.Error("Database connection lost") // OK

	// Плохие примеры (должны быть найдены линтером)
	slog.Info("User login successful!")       // ERROR
	slog.Info("User login successful 🎉")      // ERROR
	slog.Info("User login Successful")        // ERROR
	slog.Info("Пользователь вошел в систему") // ERROR
	slog.Info("Password: secret123")          // ERROR

	// Zap примеры
	logger := zap.NewExample()
	logger.Info("Request processed")     // OK
	logger.Error("Something went wrong") // OK
}
