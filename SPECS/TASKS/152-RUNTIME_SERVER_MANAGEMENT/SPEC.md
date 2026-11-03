# SPEC: 152 — RUNTIME_SERVER_MANAGEMENT

**Фича:** [019-RUNTIME_SERVER_MANAGEMENT](../../FEATURES/019-RUNTIME_SERVER_MANAGEMENT/FEATURE.md)

| Поле | Значение |
|------|----------|
| Тип | F (feature) — runtime management inbound users и WireGuard/AmneziaWG peers через Clash REST API |
| Статус | O (open) |
| Ветка | `lx` |
| Build-tag | Отдельный тег не нужен: API находится внутри существующего `with_clash_api`; protocol runtime methods без вызова не меняют поведение |

## 1. Контракт

Серверная консоль управляет работающим sing-box-lx через существующий Clash API listener и Bearer secret:

- `PUT /runtime/inbounds/{tag}/users` атомарно заменяет полный список Hysteria2/Naive users.
- `GET /runtime/inbounds/{tag}/users` возвращает stable ID и username, никогда пароль.
- `PUT /runtime/endpoints/{tag}/peers` атомарно заменяет полный список WG/AWG peers.
- `GET /runtime/endpoints/{tag}/peers` возвращает public peer data, никогда private key или PSK.

Только server-side CLI/router/desktop builds. Изменения персистентны только во внешнем источнике истины; ядро не хранит их через restart.

## 2. Семантика

- Requests адресуют runtime object по tag.
- `PUT` — полная замена и идемпотентен.
- Некорректный snapshot валидируется целиком до изменения работающего объекта.
- Пустой users list запрещён; peers list может быть пустым для полного удаления peers.
- Новые подключения/handshakes используют новую версию таблицы. Уже установленные соединения не рвутся.
- Runtime mutation не меняет listener, TLS, ключ сервера, адрес интерфейса или device-level AWG obfuscation.
- При ошибке все runtime state и маршрутизация остаются прежними.
- Пароли/PSK не возвращаются и не логируются.

## 3. HTTP contract

`PUT /runtime/inbounds/{tag}/users`

```json
{"users":[{"id":"client-1","username":"alice","password":"secret"}]}
```

`PUT /runtime/endpoints/{tag}/peers`

```json
{"peers":[{"public_key":"<base64>","pre_shared_key":"<base64>","allowed_ips":["10.0.0.2/32"],"address":"vpn.example","port":51820,"persistent_keepalive_interval":25}]}
```

Статусы: 400 invalid JSON/validation, 401 missing/invalid Clash secret, 404 unknown tag, 409 unsupported type or unavailable runtime state, 500 runtime failure.

## 4. Acceptance

1. Hysteria2 replacement использует имеющийся `hysteria2.Service.UpdateUsers`; metadata username mapping race-free.
2. Naive TCP и QUIC auth атомарно читают один immutable authenticator snapshot.
3. WG/AWG peer replacement обновляет transport peers и allowed-IP lookup без restart TUN/device.
4. Runtime endpoint отвергает mutation во время close, sleep/teardown или отсутствующего устройства.
5. HTTP routes наследуют Clash authentication.
6. GET исключает пароли, PSK и private keys.
7. Тесты покрывают auth replacement, WG routing transitions, HTTP status/auth/secret filtering и race-sensitive swaps.
8. Когда Clash API не включён, новых listener’ов и control ports не появляется; существующий Clash API не меняет поведение.

## 5. Границы

Server-side desktop/router binary через Clash REST API. Без libbox, Android, protobuf, gRPC, `lxd` admin API или нового listener. ShadowVPN2 хранит authoritative snapshot.

## 6. Upstream seams

Новая логика располагается в новых `*_lx.go` файлах. Общие adapter interfaces не расширяются; protocol packages реализуют optional capabilities. Upstream-файлы трогать только при невозможности подключить capability без минимального маркерного блока.
