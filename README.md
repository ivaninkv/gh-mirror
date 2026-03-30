# Git Mirror

CLI-инструмент для зеркалирования репозиториев между GitHub, GitVerse, GitLab и другими git-платформами.

## Возможности

- Синхронизация всех веток и тегов
- Сохранение настроек приватности репозитория
- Поддержка множественных destination-платформ
- Синхронизация по требованию
- Поддержка создания новых и обновления существующих репозиториев
- Сравнение репозиториев между source и destination
- Расширяемая архитектура для добавления новых платформ

## Требования

- Go 1.26+
- PAT (Personal Access Token) для source-платформы
- API-токен(ы) для destination-платформ(ы)

## Установка

```bash
go install ./cmd/mirror/
```

Или собрать вручную:

```bash
go build -ldflags="-s -w" -o mirror ./cmd/mirror/
```

## Настройка

Создайте `config.yaml`:

```yaml
platforms:
  github:
    token: "${GITHUB_TOKEN}"
  gitverse:
    token: "${GITVERSE_TOKEN}"
    base_url: "https://gitverse.ru/api/v1"

source: github
destinations:
  - gitverse

sync:
  timeout_minutes: 30
```

Переменные окружения поддерживаются через синтаксис `${VAR_NAME}`.

### Множественные destinations

```yaml
platforms:
  github:
    token: "${GITHUB_TOKEN}"
  gitverse:
    token: "${GITVERSE_TOKEN}"
    base_url: "https://gitverse.ru/api/v1"
  gitlab:
    token: "${GITLAB_TOKEN}"
    base_url: "https://gitlab.com/api/v4"

source: github
destinations:
  - gitverse
  - gitlab
```

## Использование

```bash
CONFIG_PATH=/path/to/config.yaml mirror sync        # синхронизация всех репозиториев
CONFIG_PATH=/path/to/config.yaml mirror sync repo   # синхронизация конкретного репозитория
CONFIG_PATH=/path/to/config.yaml mirror list        # список репозиториев source-платформы
CONFIG_PATH=/path/to/config.yaml mirror diff        # различия между source и первым destination
```

## Настройка GitHub Token

1. Перейдите в GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Создайте новый токен:
   - Resource owner: ваше имя пользователя
   - Repository access: All repositories
   - Permissions: Contents → Read and write
3. Скопируйте токен

## Настройка GitVerse Token

1. Перейдите в GitVerse → Profile → Settings → Tokens
2. Создайте новый токен
3. Скопируйте токен

## Добавление новой платформы

1. Создайте `pkg/platform/<platform_name>/client.go`
2. Реализуйте интерфейс `Platform`:

```go
type Platform interface {
    ID() models.PlatformID  // уникальный идентификатор, напр. "gitlab"
    Name() string          // человекочитаемое имя, напр. "GitLab"
    Configure(token string, baseURL string) error

    GetAuthenticatedUser(ctx context.Context) (string, error)
    ListRepositories(ctx context.Context) ([]models.Repository, error)
    GetRepository(ctx context.Context, owner, repo string) (*models.Repository, error)
    CreateRepository(ctx context.Context, name string, private bool, description string) (*models.Repository, error)
    UpdateRepository(ctx context.Context, owner, repo string, private bool, description string) error
    RepositoryExists(ctx context.Context, owner, repo string) (bool, error)
    CloneURL(repo models.Repository, token string) string
}
```

3. Зарегистрируйте платформу в `init()`:

```go
func init() {
    platform.Register("gitlab", func() platform.Platform {
        return &GitLabClient{}
    })
}
```

4. Добавьте конфигурацию в `config.yaml`:

```yaml
platforms:
  gitlab:
    token: "${GITLAB_TOKEN}"
    base_url: "https://gitlab.com/api/v4"
```

## Как это работает

1. Получает список всех репозиториев с source-платформы (включая приватные)
2. Для каждого репозитория и каждого destination:
   - Проверяет существование репозитория на destination
   - Создаёт или обновляет метаданные (приватность, описание)
   - Клонирует репозиторий с помощью go-git (pure Go)
   - Push-ит все ветки и теги
3. Логирует репозитории, которые есть на destination, но отсутствуют на source (никогда не удаляет)

## Лицензия

MIT
