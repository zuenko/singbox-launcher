# План: лаунчер «Не отвечает» после сна/гибернации

Реализация событий питания — в **internal/platform** (Windows: power_windows.go; прочие ОС: power_stub.go). Клиенты не проверяют GOOS: вызывают platform API всегда; платформа сама решает, слать события или no-op.

**Контракт platform (см. SPEC, раздел «Архитектура: события питания»):** подписка на события **sleep** / **resume** и читаемый **статус sleep**. Подписчики используют проверку статуса и контекст для прерывания запросов и не стартуют новую работу при sleep.

## Компоненты и изменения

1. **api/clash.go**
   - Вынос создания HTTP-клиента в `clashHTTPClient()` с заданным `IdleConnTimeout` (30 с).
   - Глобальный клиент защищён мьютексом; `getHTTPClient()` возвращает текущий клиент для всех запросов.
   - `ResetClashHTTPTransport()` создаёт новый клиент, подменяет глобальный, закрывает idle-соединения старого транспорта.
   - Все вызовы `httpClient.Do` заменены на `getHTTPClient().Do`.
   - Пакет api импортирует platform; внутри — хелперы `requestContext()` (IsSleeping → ErrPlatformInterrupt, иначе PowerContext()) и `normalizeRequestError()` (context.Canceled → ErrPlatformInterrupt). Все HTTP-функции используют их; публичный API без контекста, при sleep/отмене возвращают ErrPlatformInterrupt.

2. **internal/platform (Windows)**
   - `power_windows.go`: скрытое окно и цикл сообщений в отдельной горутине с LockOSThread; WM_POWERBROADCAST (sleep 4, resume 7/18); IsSleeping(), PowerContext(), RegisterSleepCallback, RegisterPowerResumeCallback, StopPowerResumeListener. Документация API без привязки к «только Windows» — платформа решает.

3. **internal/platform (не Windows)**
   - `power_stub.go`: IsSleeping() всегда false, PowerContext() — context.Background(); RegisterSleepCallback, RegisterPowerResumeCallback, StopPowerResumeListener — no-op.

4. **main.go**
   - После `UpdateUI()` безусловный вызов `platform.RegisterPowerResumeCallback` с callback (ResetClashHTTPTransport + лог). На не-Windows — no-op внутри platform.
   - В блоке очистки перед `GracefulExit` — безусловный вызов `platform.StopPowerResumeListener()`.

## Файлы

- `api/clash.go` — IdleConnTimeout, mutex, getHTTPClient, ResetClashHTTPTransport; requestContext, normalizeRequestError; все HTTP-функции через них.
- `internal/platform/power_windows.go` — build tag windows; реализация событий питания.
- `internal/platform/power_stub.go` — build tag !windows; no-op.
- `main.go` — безусловные вызовы RegisterPowerResumeCallback и StopPowerResumeListener (без проверки GOOS).
