# 019 — Runtime server management

## Назначение

Внешняя серверная консоль может менять учётные данные inbound-пользователей и WG/AWG peers работающего ядра без перезапуска процесса. Полная конфигурация при следующем запуске остаётся источником истины.

## Контролируемые операции

| Операция | HTTP | Тело |
|---|---|---|
| Заменить пользователей inbound | `PUT /runtime/inbounds/{tag}/users` | `{ "users": [{ "id": "...", "username": "...", "password": "..." }] }` |
| Получить пользователей inbound | `GET /runtime/inbounds/{tag}/users` | — |
| Заменить peers endpoint | `PUT /runtime/endpoints/{tag}/peers` | `{ "peers": [...] }` |
| Получить peers endpoint | `GET /runtime/endpoints/{tag}/peers` | — |

Все операции требуют существующий Clash API Bearer secret. `PUT` заменяет полный snapshot и идемпотентен. Ответы не содержат пароли или private keys.

## Входы и выходы

Входы: HTTP-запросы локальной консоли, inbound/endpoint tags и полные snapshots credentials/peers.

Выходы: JSON со стабильным tag и числом применённых записей; HTTP ошибки на некорректный запрос, неизвестный tag, неподдерживаемый тип или отказ runtime-операции.

## Data flow

HTTP handler проверяет запрос и tag, находит объект через manager, проверяет runtime capability и передаёт snapshot в protocol implementation. Новые подключения используют обновлённые credentials; существующие соединения сохраняются. WG/AWG получает новую таблицу peers и маршрутов без пересоздания TUN-интерфейса.

## Гарантии

- Все runtime routes наследуют существующие Clash API authentication и CORS middleware.
- Неизвестный tag не создаёт объект.
- Пустой snapshot пользователей отвергается, чтобы не оставить серверный inbound без доступа.
- Duplicate IDs/usernames/public keys, пустые обязательные значения и некорректные адреса отвергаются до публикации.
- При ошибке применения прежний runtime snapshot остаётся активен.
- Пароли и PSK не возвращаются из GET и не пишутся в логи.
- Операции не сохраняются ядром: оператор повторно отправляет snapshot из своего источника истины после перезапуска.
- Полная смена listener/TLS/AWG device-level settings остаётся обычным config reload.

## Границы

Только server-side desktop/router binary через Clash REST API. Без libbox, Android, `StartedService`, protobuf, gRPC и V2Ray API. Не является постоянным хранилищем пользователей.

## Особенности сопровождения

Runtime capabilities остаются в protocol packages; Clash handlers только разбирают запрос и маршрутизируют вызов. Общие upstream interfaces не расширяются. Изменения upstream-файлов ограничиваются минимальными `// lx:` seams.
