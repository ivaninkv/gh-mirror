# Git Mirror

CLI-инструмент для зеркалирования репозиториев между GitHub, GitLab, GitVerse, Codeberg и другими git-платформами.

## Возможности

- Синхронизация всех веток и тегов
- Сохранение настроек приватности репозитория
- Поддержка множественных destination-платформ
- Синхронизация по требованию
- Поддержка создания новых и обновления существующих репозиториев
- Сравнение репозиториев между source и destination

## Скачивание

Перейдите на страницу [Releases](https://github.com/ivaninkv/gh-mirror/releases/latest) и скачайте архив для вашей операционной системы:

| ОС | Архитектура | Файл |
|-----|-------------|------|
| Windows | x86_64 | `mirror-windows-amd64.zip` |
| Linux | x86_64 | `mirror-linux-amd64.tar.gz` |
| Linux | ARM64 | `mirror-linux-arm64.tar.gz` |
| macOS | Intel | `mirror-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `mirror-darwin-arm64.tar.gz` |

## Установка

Распакуйте архив:

```bash
# Linux / macOS
tar -xzf mirror-*-*.tar.gz

# Windows
unzip mirror-windows-amd64.zip
```

После распаковки вы получите исполняемый файл `mirror` (или `mirror.exe` для Windows).

## Настройка

Создайте файл `config.yaml` в одном из следующих расположений (в порядке приоритета):

1. `./config.yaml` (рядом с исполняемым файлом)
2. `~/.config/gh-mirror/config.yaml`
3. `/etc/gh-mirror/config.yaml`

Пример `config.yaml`:

```yaml
platforms:
  github:
    token: "${GITHUB_TOKEN}"
    url: "https://github.com"
  gitverse:
    token: "${GITVERSE_TOKEN}"
    api_url: "https://api.gitverse.ru"
    url: "https://gitverse.ru"
  codeberg:
    token: "${CODEBERG_TOKEN}"
    api_url: "https://codeberg.org/api/v1"
    url: "https://codeberg.org"

source: github
destinations:
  - gitverse
  - codeberg

sync:
  timeout_minutes: 30
```

Токены передаются через переменные окружения:

```bash
export GITHUB_TOKEN="your-github-token"
export GITVERSE_TOKEN="your-gitverse-token"
export CODEBERG_TOKEN="your-codeberg-token"
```

### Множественные destinations

```yaml
platforms:
  github:
    token: "${GITHUB_TOKEN}"
    url: "https://github.com"
  gitverse:
    token: "${GITVERSE_TOKEN}"
    api_url: "https://api.gitverse.ru"
    url: "https://gitverse.ru"
  gitlab:
    token: "${GITLAB_TOKEN}"
    api_url: "https://gitlab.com/api/v4"
    url: "https://gitlab.com"
  codeberg:
    token: "${CODEBERG_TOKEN}"
    api_url: "https://codeberg.org/api/v1"
    url: "https://codeberg.org"

source: github
destinations:
  - gitverse
  - gitlab
  - codeberg
```

## Использование

```bash
./mirror sync                        # синхронизация всех репозиториев
./mirror sync owner/repo             # синхронизация конкретного репозитория
./mirror list                        # список репозиториев source-платформы
./mirror diff                        # различия между source и первым destination
./mirror --help                      # справка
./mirror --version                   # версия программы
```

Если `config.yaml` находится в стандартном расположении, указывать `CONFIG_PATH` не нужно. Для указания другого пути:

```bash
CONFIG_PATH=/path/to/config.yaml ./mirror sync
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

## Настройка GitLab Token

1. Перейдите в GitLab → Profile → Access Tokens
2. Создайте новый токен:
   - Token name: произвольное имя
   - Expiration date: без ограничения срока
   - Scopes: `read_api`, `write_repository`
3. Скопируйте токен

## Настройка Codeberg Token

1. Перейдите в Codeberg → Settings → Applications
2. Создайте новый токен:
   - Token name: произвольное имя
   - Scopes: `write:repo` (required for creating repositories)
3. Скопируйте токен

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
