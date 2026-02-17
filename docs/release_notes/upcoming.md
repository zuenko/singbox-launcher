# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

Ниже — всё смысловое из коммитов с момента v0.8.1 (с 8958ec93 и далее).

---

## EN

### Highlights
- **Windows:** Start/stop is more stable. Graceful stop via console (AttachConsole + CTRL_BREAK) when the core has a console; fallback to taskkill. taskkill tries without `/F` first, then with `/F` on error (same for Kill by name in Help tab). Fewer crashes and WinTun issues on restart.
- **Logging:** Default log level is now build-based (release = Warn, dev = Verbose). Removed `SINGBOX_DEBUG` env var. API log level follows global level; `logFile` param removed from API methods. Less noise: core version and stability timer are cached.
- **Config:** `getConfigJSON` outputs trailing commas for all config readers. Windows TUN: removed netsh cleanup on stop (interfaces close normally).

---

## RU

### Основное
- **Windows:** Запуск и остановка работают стабильнее. Мягкая остановка по консоли (AttachConsole + CTRL_BREAK), при необходимости — fallback на taskkill. taskkill сначала без `/F`, при ошибке — с `/F` (так же для Kill по имени во вкладке Help). Меньше крашей и проблем с WinTun при перезапуске.
- **Логирование:** Уровень логов по умолчанию зависит от сборки (release = Warn, dev = Verbose). Убрана переменная `SINGBOX_DEBUG`. Уровень api.log следует глобальному; параметр `logFile` убран из методов API. Меньше шума: кэшируются версия ядра и таймер стабильности.
- **Конфиг:** `getConfigJSON` выводит trailing commas для всех читателей конфига. Windows TUN: убрана очистка через netsh при остановке (интерфейсы закрываются сами).
