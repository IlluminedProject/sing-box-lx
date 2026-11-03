# PLAN: 152 — RUNTIME_SERVER_MANAGEMENT

Контракт — [SPEC.md](SPEC.md).

## Этап 1. Runtime capabilities

1. Добавить optional `adapter.RuntimeUserManager` и `adapter.RuntimePeerManager` в новые lx-файлы; не менять базовые `Inbound`/`Endpoint` interfaces.
2. Hysteria2: wrapper над `Service.UpdateUsers`; immutable mapping user ID → username публиковать атомарно, чтобы существующие TCP/UDP callbacks не читали изменяемый slice.
3. Naive: atomic pointer на `auth.Authenticator`; инициализация и TCP/QUIC auth используют общий новый user table.
4. WG/AWG protocol endpoint: сериализовать mutation с suspend/teardown/close, формировать peer delta в формате IpcSet; синхронно обновить transport peers и allowed-IP route lookup; не пересоздавать TUN.
5. Проверить IpcSet remove/update semantics в включённом submodule. Если IpcSet не позволяет надёжно удалить peer или заменить allowed IPs, остановить эту часть и оформить minimal isolated submodule patch с red/green тестом.

## Этап 2. Clash REST

1. Добавить `experimental/clashapi/runtime_management_lx.go` и тесты.
2. Смонтировать `/runtime` внутри существующей authenticated chi group в `server.go`; отдельного listener не создавать.
3. Получить `adapter.InboundManager` и `adapter.EndpointManager` из service context; найти по tag, вызвать optional capability.
4. Полностью валидировать DTO перед mutation; не логировать credentials; GET response redacts password/PSK/private key.
5. API доступно только в сборках с `with_clash_api`; без него нет дополнительного build/tag wiring.

## Этап 3. ShadowVPN2 integration

1. Добавить typed `SingBoxRuntimeClient`, configured `BaseAddress` + secret, loopback-only default.
2. После успешного RavenDB save отправлять snapshots users/peers по существующим inbound/endpoint tags.
3. Сериализовать операции клиента с full config regeneration через общий gate. При runtime failure использовать full-config recovery path и явно отдавать/логировать application failure по текущей модели.
4. Не трогать независимые пользовательские изменения в `WireGuardConnectionInfo.cs`.

## Этап 4. Проверки

- `go test -race` для затронутых protocol, adapter и Clash API пакетов.
- `go vet` затронутых пакетов.
- ShadowVPN2 unit tests для request serialization, HTTP error handling и update ordering.
- Smoke: поднять sing-box test config с loopback Clash API, применить snapshot, проверить auth/GET/PUT. WG peer replacement проверить device IPC и route lookup без публичного сервера.
- Обновить FEATURE, feature index, Roadmap, TASKS.md, IMPLEMENTATION_REPORT.md.
