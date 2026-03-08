# Отчёт о реализации: 001 — лаунчер «Не отвечает» после сна/гибернации

**Статус:** реализовано (resume + сброс транспорта; sleep/resume события и подписчики; рефакторинг без зависимости клиентов от ОС).

**Дата:** 2025-03.

## Изменения

1. **Сброс HTTP-транспорта Clash API после resume**  
   При событии resume платформа вызывает зарегистрированный callback; в main зарегистрирован callback, вызывающий `api.ResetClashHTTPTransport()` и логирование. На Windows — по WM_POWERBROADCAST (resume 7/18); на прочих ОС callback не вызывается (no-op).

2. **События sleep/resume и статус sleep (platform)**  
   - Sleep (на Windows — wParam 4): отмена PowerContext(), установка статуса sleep, вызов RegisterSleepCallback; логирование «power: system entering sleep/hibernation».
   - Resume (7/18): сброс sleep, новый PowerContext(), вызов resume callback; логирование «power: system resumed from sleep/hibernation».
   - **IsSleeping()**, **PowerContext()**, **RegisterSleepCallback**, **RegisterPowerResumeCallback**, **StopPowerResumeListener** — единый API; документация без привязки к «только Windows» (платформа сама решает, слать события или no-op).

3. **api/clash.go**  
   - IdleConnTimeout 30 с; getHTTPClient() под мьютексом; ResetClashHTTPTransport().
   - Хелперы **requestContext()** (при IsSleeping возвращает ErrPlatformInterrupt, иначе PowerContext()) и **normalizeRequestError()** (context.Canceled → ErrPlatformInterrupt). TestAPIConnection, GetProxiesInGroup, SwitchProxy, GetDelay используют их; публичный API без контекста, при sleep/отмене возвращают **ErrPlatformInterrupt**. Зависимости от runtime.GOOS в api нет.

4. **Клиенты без зависимости от ОС**  
   main.go вызывает `RegisterPowerResumeCallback` и `StopPowerResumeListener` безусловно (без проверки GOOS). Остальные клиенты (api_service, clash_api_tab) только проверяют platform.IsSleeping() и обрабатывают api.ErrPlatformInterrupt. Реализация событий — только в platform (power_windows.go / power_stub.go).

## Риски и ограничения

- **Блокировка UI на ~7 минут** по гипотезе может быть в драйвере/OpenGL; сброс соединений и обработка resume не гарантируют устранение этой блокировки, но снижают нагрузку и последствия «протухших» TCP после пробуждения.
- Полная сборка (`go build .`) на машине без CGO/компилятора C (для go-gl) по-прежнему может падать; к изменённым пакетам это не относится.

## Проверка

- `go build ./api/... ./internal/platform/...` — успешно (Windows).
- `go vet ./api/ ./internal/platform/` — без замечаний.
- Вручную: запуск лаунчера на Windows, переход в сон/гибернацию, пробуждение — в логе должно появиться сообщение «Power resume: Clash API HTTP transport reset».
