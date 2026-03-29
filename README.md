# GitHub → GitVerse Mirror

CLI-инструмент для зеркалирования репозиториев с GitHub (включая приватные) в GitVerse с сохранением всех веток и тегов.

## Возможности

- Синхронизация всех веток и тегов
- Сохранение настроек приватности репозитория
- Синхронизация по требованию
- Поддержка создания новых и обновления существующих репозиториев
- Сравнение репозиториев между GitHub и GitVerse

## Требования

- Go 1.26+
- Git
- GitHub Personal Access Token с правами `repo`
- API-токен GitVerse

## Установка

```bash
go build -o mirror ./cmd/mirror/
```

## Настройка

Создайте `config.yaml`:

```yaml
github:
  token: "${GITHUB_TOKEN}"  # PAT с правами repo

gitverse:
  token: "${GITVERSE_TOKEN}"
  base_url: "https://gitverse.ru/api/v1"

sync:
  timeout_minutes: 30
```

Переменные окружения поддерживаются через синтаксис `${VAR_NAME}`.

## Использование

### Синхронизация всех репозиториев

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror sync
```

Или с файлом конфигурации:

```bash
CONFIG_PATH=/path/to/config.yaml ./mirror sync
```

### Синхронизация конкретного репозитория

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror sync repository-name
```

### Список репозиториев на GitHub

```bash
GITHUB_TOKEN=ghp_xxx ./mirror list
```

### Показать различия между GitHub и GitVerse

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror diff
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

## Как это работает

1. Получает список всех репозиториев с GitHub (включая приватные)
2. Для каждого репозитория:
   - Создаёт или обновляет репозиторий на GitVerse (с сохранением приватности)
   - Использует `git clone --mirror` для полного клонирования репозитория
   - Использует `git push --mirror` для отправки всех веток и тегов в GitVerse
3. Логирует репозитории, которые есть на GitVerse, но отсутствуют на GitHub (никогда не удаляет)

## Лицензия

MIT
