# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

Ниже — всё смысловое из коммитов с момента v0.8.1 (с 8958ec93 и далее).

---

## EN

### Highlights
- **Unified “download failed” dialog** — When download of sing-box, wintun.dll, config template, or SRS fails, one dialog is shown: short message, “Open download page” link with a copy-URL button, “Open folder”, and “Close”. Same behavior for all resources.
- **SRS local download** — Rule-set files (SRS) from `raw.githubusercontent.com` are now downloaded locally to `bin/rule-sets/`. SRS button (⬇/🔄/✔️) in Rules tab; tooltip on hover shows original URL from template. Removed `go-any-way-githubusercontent` outbound and `download_detour`/`update_interval` from rule_set.
- **Windows:** Start/stop is more stable. Graceful stop via console (AttachConsole + CTRL_BREAK) when the core has a console; fallback to taskkill. taskkill tries without `/F` first, then with `/F` on error (same for Kill by name in Help tab). Fewer crashes and WinTun issues on restart.
- **Logging:** Default log level is now build-based (release = Warn, dev = Verbose). Removed `SINGBOX_DEBUG` env var. API log level follows global level; `logFile` param removed from API methods. Less noise: core version and stability timer are cached.
- **Ping error tooltip (Clash API)** — Ping errors on the Servers tab are now shown directly in a tooltip on the Ping button (no modal dialog). Tooltips use `fyne-tooltip` both in the main window and in the config wizard; the old custom tooltip wrapper has been removed.
- **Ping test concurrency (Clash API)** — The `test` button on the Servers tab now pings proxies in parallel with a limited number of concurrent requests (20 by default), making full-list ping tests noticeably faster while keeping Clash API load under control. Errors from these tests are also reflected in Ping button tooltips.
- **Ping test endpoints & pinning (Clash API)** — Ping delay endpoint is now configurable from the UI (GStatic, Google, Gosuslugi, YaStatic, or a custom URL). In the proxy list, `direct-out` (if present) and the currently active proxy are always pinned to the top, regardless of sort order.
- **Config:** `getConfigJSON` outputs trailing commas for all config readers. Windows TUN: removed netsh cleanup on stop (interfaces close normally).

---

## RU

### Основное
- **Единый диалог «загрузка не удалась»** — при ошибке загрузки sing-box, wintun.dll, шаблона конфига или SRS показывается один диалог: короткое сообщение, ссылка «Open download page» с кнопкой копирования URL, «Open folder» и «Close». Одинаковое поведение для всех ресурсов.
- **Локальное скачивание SRS** — rule-set файлы с `raw.githubusercontent.com` скачиваются локально в `bin/rule-sets/`. Кнопка SRS (⬇/🔄/✔️) во вкладке Rules; при наведении — tooltip с оригинальным URL из шаблона. Удалены outbound `go-any-way-githubusercontent` и `download_detour`/`update_interval` из rule_set.
- **Windows:** Запуск и остановка работают стабильнее. Мягкая остановка по консоли (AttachConsole + CTRL_BREAK), при необходимости — fallback на taskkill. taskkill сначала без `/F`, при ошибке — с `/F` (так же для Kill по имени во вкладке Help). Меньше крашей и проблем с WinTun при перезапуске.
- **Логирование:** Уровень логов по умолчанию зависит от сборки (release = Warn, dev = Verbose). Убрана переменная `SINGBOX_DEBUG`. Уровень api.log следует глобальному; параметр `logFile` убран из методов API. Меньше шума: кэшируются версия ядра и таймер стабильности.
- **Tooltip ошибки Ping (Clash API)** — во вкладке Servers ошибка Ping теперь показывается прямо в tooltip на кнопке Ping (без отдельного модального диалога). Tooltips реализованы через `fyne-tooltip` и в главном окне, и в визарде конфигурации; старая самодельная обёртка tooltip удалена.
- **Параллельный Ping test (Clash API)** — кнопка `test` во вкладке Servers теперь пингует прокси параллельно, с ограничением числа одновременных запросов (по умолчанию 20), что заметно ускоряет полную проверку списка и не перегружает Clash API. Ошибки из этих тестов также попадают в tooltip кнопок Ping.
- **Ping endpoints и закрепление сверху (Clash API)** — endpoint проверки Ping теперь настраивается из UI (GStatic, Google, Gosuslugi, YaStatic или произвольный URL). В списке прокси `direct-out` (если есть) и текущий активный прокси всегда закреплены вверху, независимо от выбранной сортировки.
- **Конфиг:** `getConfigJSON` выводит trailing commas для всех читателей конфига. Windows TUN: убрана очистка через netsh при остановке (интерфейсы закрываются сами).
