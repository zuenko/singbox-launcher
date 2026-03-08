# Задачи: 001 — лаунчер «Не отвечает» после сна/гибернации

- [x] api/clash.go: IdleConnTimeout (30 с), clashHTTPClient(), mutex + getHTTPClient(), ResetClashHTTPTransport(), заменить все httpClient.Do на getHTTPClient().Do
- [x] internal/platform/power_windows.go: скрытое окно, WM_POWERBROADCAST (resume 7/18, sleep 4), RegisterPowerResumeCallback, RegisterSleepCallback, StopPowerResumeListener; состояние sleep, PowerContext(), IsSleeping()
- [x] internal/platform/power_stub.go: IsSleeping false, PowerContext background, RegisterSleepCallback no-op
- [x] main.go: безусловно RegisterPowerResumeCallback (ResetClashHTTPTransport + лог) и StopPowerResumeListener; в таймере обновления трея проверка IsSleeping()
- [x] api/clash.go: requestContext(), normalizeRequestError(); HTTP-функции используют их, возвращают ErrPlatformInterrupt при sleep/отмене; контекст не в публичном API; без зависимости от GOOS
- [x] core/services/api_service.go, ui/clash_api_tab.go: проверка platform.IsSleeping() перед запросами (ранний выход)
- [x] PLAN.md, IMPLEMENTATION_REPORT.md обновлены
