# LogLinter

Статический анализатор , проверяющий сообщения логов на соответствие правилам стиля и безопасности.

## Возможности

- Проверяет, начинаются ли сообщения журнала с маленькой буквы
- Обеспечивает отображение только сообщений на английском языке
- Обнаруживает специальные символы и эмодзи
- Выявляет потенциально конфиденциальные данные
- Работает с несколькими системами логирования (log, slog, zap)
- **Настраиваемые правила** через флаги командной строки и файл конфигурации
- **Автоматическая коррекция** с помощью SuggestedFixes для нарушений, связанных с использованием строчных букв
- **Поддержка пользовательских конфиденциальных ключевых слов**

## Правила

### Lowercase rule
```go
slog.Info("user logged in")        // ✅ OK
slog.Info("User logged in")        // ❌ start with lowercase
```

### English-only rule
```go
slog.Info("user authenticated")    // ✅ OK
slog.Info("пользователь вошёл")     // ❌ not English only
```

### Special characters rule
```go
slog.Info("user logged in")        // ✅ OK
slog.Info("user logged in.")       // ❌ should not contain special characters
slog.Info("🔥 fire")                // ❌ should not contain emoji
```

### Sensitive data rule
```go
slog.Info("user request processed") // ✅ OK
slog.Info("password: secret123")     // ❌ may contain sensitive data
slog.Info("login: " + username)      // ❌ concatenates potentially sensitive variable
```

## Поддерживаемые логгеры

- `log.Info()`, `log.Error()`, `log.Warn()`, `log.Debug()`, `log.Fatal()`, `log.Panic()`
- `log.Print()`, `log.Printf()`, `log.Println()`
- `slog.Info()`, `slog.Error()`, `slog.Warn()`, `slog.Debug()`
- `zap.Info()`, `zap.Error()`, `zap.Warn()`, `zap.Debug()`

## Встроенные "чувствительные" слова

- `password`, `passwd`, `secret`, `token`
- `api_key`, `apikey`, `apitoken`, `auth`
- `credential`, `private_key`, `privatekey`

*Расширяемо через конфигурацию.*

## Установка

### Как плагин для golangci-lint

1. Склонировать репозиторий
``` bash
git clone git@github.com:golchanskiy23/LogLinter.git
```
2. Перейти в репозиторий
``` bash
cd ./LogLint
```
3. Собрать плагин и перейти в рабочую директорию:
```bash
go build -buildmode=plugin -o loglint.so ./plugin
cd ../your_dir
```

4. Добавить в свой проект `.golangci.yml`:
```yaml
linters-settings:
  custom:
    loglint:
      path: ../LogLint/loglint.so
      description: "checks log messages for style and security"
      original-url: github.com/golchanskiy23/loglint

linters:
  enable:
    - loglint
```

5. Запуск golangci-lint:
```bash
golangci-lint run
```

### Локальный бинарник

```bash
# 1. Склонировать репозиторий и собрать бинарник
git clone git@github.com:golchanskiy23/LogLinter.git
cd ../LogLinter
go build -o loglint ./cmd/loglint

# 2. Запустить на рабочем проекте
./loglint ../your_project/...
```

## Конфигурация

LogLinter поддерживает гибкую конфигурацию через **configuration file** флаги командной строки.

### Конфигурационный файл**

Создайте `.loglint.yml` в корне проекта:

```yaml
# LogLinter Configuration File

# Disable specific rules
disable-lowercase: false
disable-english: false
disable-special: false
disable-sensitive: false

# Add custom sensitive keywords
extra-sensitive:
  - "custom_secret"
  - "private_data"
  - "user_token"
```

### Флаги командной строки (только через бинарник)

```bash
./loglint -disable-lowercase ./...
./loglint -extra-sensitive="custom_secret,private_data" ./...
```

### Автокоррекция с SuggestedFixes

LogLinter использует **SuggestedFixes** для автоматической коррекции значений кода в IDE:

```go
// Before:
log.Print("Error occurred while processing")

// Auto-corrected to:
log.Print("error occurred while processing")
```

**Использование:**
```bash
# Apply fixes automatically
./loglint -fix ./...
```

**Поддержка правил:**
- Автокоррекция Lowercase rule
