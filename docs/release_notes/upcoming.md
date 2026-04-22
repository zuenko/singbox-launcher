# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

---

## EN

### Highlights

- Subscriptions: **Xray/V2Ray JSON array** body (`[ { full config }, … ]`) — one logical node per element; **`dialerProxy`** (or **`dialer`**) to a **SOCKS** or **VLESS** outbound → sing-box **`detour`** (jump outbound emitted first). Non-empty **`remarks`** → **`Label`** (full text) and tags **`{slug}`** / **`{slug}_jump_server`** for main vs jump (else `xray-{i}` / `xray-{i}_jump_server`); slug keeps letters/digits and **Unicode regional indicators** (UTF flag emoji), then usual prefix/unique rules. Example: `docs/examples/xray_subscription_array_sample.json`. Share URI: outbounds with **`detour`** are not encodable (**`ErrShareURINotSupported`**).
- **VLESS:** no longer auto-fills **`flow: xtls-rprx-vision`** when **`flow`** is missing in the URI or in Xray JSON — add **`flow`** in the subscription if the server requires Vision.

### Technical / Internal

- **Licensing:** the project is now under **GPL-3.0** with optional commercial licensing from Leadaxe (see `LICENSING.md`); previously MIT.

### Added — Mobile parity pass (2026-04-22 night-work)

- **Debug API (localhost-only):** optional HTTP server on `127.0.0.1:9269`. Off by default. Endpoints `/ping`, `/version`, `/state`, `/proxies`, `/action/{start,stop,update-subs}`. Bearer-token auth (32-byte random), generated on first enable, surfaced via Copy-token button on the Diagnostics tab. Ported from LxBox spec 031, trimmed for desktop (no rules/subs CRUD, no `/config` dump, no `/logs` tail).
- **Auto-ping after VPN connect (default ON):** 5 s after sing-box enters the running state, the Servers tab auto-runs Test on all proxies so latency is fresh when the user switches tabs. Toggle on Core Dashboard, persisted in `settings.json`.
- **Subscription auto-update global toggle:** dedicated checkbox on Core Dashboard (default ON). Off skips all scheduled refreshes; manual Update always works. Re-enabling from UI also resets the auto-disable-on-failure counter.
- **URLTest parameters as template vars with preset dropdowns:** `auto-proxy-out.url / interval / tolerance` hoisted to `@urltest_url` / `@urltest_interval` / `@urltest_tolerance`, each with a `{title, value}` preset dropdown (Cloudflare / GStatic / Google connectivitycheck; 1m / 5m (default) / 10m / 30m / 1h; 50 / 100 / 200 / 500 ms).
- **Template `vars[].options` accepts `{title, value}` form:** user-visible labels can differ from substitution values. Legacy string-list form still works (title == value). Mixed arrays also supported.
- **Right-click menu on the Update button:** Reconnect / Restart / Start / Stop / Update-subs in one explicit popup. Primary-tap still runs Update as before.
- **Keyboard shortcuts:** `Cmd/Ctrl+R` reconnect, `Cmd/Ctrl+U` update-subs.
- **Power-resume hooks:** on wake, reset Clash API HTTP transport, schedule `RefreshAPIFunc` + auto-ping-after-connect so latency is fresh. Linux gained native support via systemd-logind `PrepareForSleep` over DBus (reuses existing indirect dep, no new deps). macOS still stubbed — IOKit cgo follow-up. Spec 011 partial closure.

### Added — Resilience & observability

- **Atomic writes** for `config.json` (scheduled parser + wizard save) and `settings.json` — stage to `.tmp` / `.swap` then rename, so a crash mid-write can't truncate the live file.
- **100 MB download cap** on sing-box core downloads (pre-flight Content-Length + mid-stream cumulative) to contain a compromised or misconfigured mirror.
- **Per-source parser summary:** success toast now reads `"Configuration updated: 2/3 source(s) succeeded (1 failed)"` when any subscription source fails or returns zero nodes.
- **Last-auto-update failure pill** on Core Dashboard — shows the actual error message (with HTTP status humanization: `401 → "token may have expired"`, `429 → "rate limited — try again later"`, etc.) for up to 24 h after a scheduled-update failure, auto-clears on next success.
- **"(subs: Xh ago)"** hint next to the config modTime on Core Dashboard.
- **Dirty-config marker (`*`)** on the Update button whenever wizard has saved template/state changes that the parser has not yet applied.

### Added — Security hygiene

- **Clash API token redaction** in debug logs — no more full-secret leak when users paste diag logs into GitHub issues.
- **Ping-all button lock** while a test is in flight (prevents duplicate workers on spam-click).

### Fixed

- **CI red since 2026-04-16:** `internal/locale.TestAllKeysPresent` missing 4 `ru.json` translations for the jump-outbound share-URI menu items.
- **STUN settings dialog** clipping on Windows (issue **#54**) — force ≥ 520 px width.

### Changed — Template defaults

- **Russian & Cyrillic TLDs expanded** in `ru-domains` rule-set: added `.рус / .москва / moscow / tatar / .дети / .сайт / .орг / .ком`.

---

## RU

### Основное

- Подписки: **JSON-массив** полных конфигов **Xray** (`[ {...}, … ]`) — по одной логической ноде на элемент; **`dialerProxy`**/**`dialer`** → hop **SOCKS** или **VLESS**, затем основной outbound с **`detour`**. **`remarks`**: полный текст в **`Label`** и в комментарии к outbound в JSON; теги: основной **`{slug}`**, jump **`{slug}_jump_server`** (или **`xray-{i}`** / **`xray-{i}_jump_server`** без `remarks`); в slug сохраняются буквы/цифры и **символы UTF-флагов** (региональные индикаторы). Пример: **`docs/examples/xray_subscription_array_sample.json`**. «Копировать ссылку» для таких нод с цепочкой — не поддерживается (**`detour`**).
- **VLESS:** больше **не подставляется** автоматически **`flow: xtls-rprx-vision`**, если в ссылке или в JSON Xray **`flow` не задан** — при необходимости Vision укажите **`flow`** в подписке.

### Техническое / Внутреннее

- **Лицензия:** репозиторий под **GPL-3.0**, при необходимости коммерческая лицензия — см. `LICENSING.md` (ранее MIT).

### Добавлено — паритет с мобильным клиентом (ночь 2026-04-22)

- **Debug API (локальный HTTP):** опциональный сервер на `127.0.0.1:9269`. По умолчанию выключен. Эндпоинты `/ping`, `/version`, `/state`, `/proxies`, `/action/{start,stop,update-subs}`. Bearer-токен (32 байта crypto/rand), генерируется при первом включении, показывается кнопкой «Копировать токен» на вкладке Диагностика. Портировано из LxBox spec 031 в урезанном виде (без CRUD правил/подписок, без `/config`, без `/logs`).
- **Автопинг после подключения VPN (по умолчанию ВКЛ):** через 5 секунд после входа sing-box в running-статус вкладка Servers автоматически пингует все прокси — когда пользователь переключится туда, задержки уже свежие. Чекбокс в Core Dashboard, сохраняется в `settings.json`.
- **Глобальный выключатель автообновления подписок:** чекбокс в Core Dashboard (по умолчанию ВКЛ). OFF — все плановые обновления пропускаются; ручной Update всегда работает. Включение вручную также сбрасывает счётчик последовательных ошибок.
- **Параметры URLTest как шаблонные vars с preset-дропдаунами:** `auto-proxy-out.url / interval / tolerance` вынесены в `@urltest_url` / `@urltest_interval` / `@urltest_tolerance` с `{title, value}` пресетами (Cloudflare / GStatic / Google connectivitycheck; 1m / 5m (default) / 10m / 30m / 1h; 50 / 100 / 200 / 500 мс).
- **`vars[].options` поддерживает форму `{title, value}`:** подписи для дропдаунов отдельно от подставляемых значений. Старая форма (массив строк) продолжает работать — title == value. Смешанные массивы тоже можно.
- **Правый клик по кнопке Update:** popup-меню Reconnect / Restart / Start / Stop / Update-subs — явный доступ ко всем операциям в один клик. Левый клик по-прежнему запускает Update.
- **Горячие клавиши:** `Cmd/Ctrl+R` — reconnect, `Cmd/Ctrl+U` — update-subs.
- **Обработка wake-from-sleep:** на резюм сбрасываются HTTP-соединения Clash API, затем `RefreshAPIFunc` и автопинг. Linux получил нативную поддержку через systemd-logind `PrepareForSleep` по DBus (использует существующий indirect-dep, новых зависимостей нет). macOS пока stub — IOKit cgo в следующий релиз. Частичное закрытие spec 011.

### Добавлено — устойчивость и наблюдаемость

- **Атомарные записи** `config.json` (парсер + визард) и `settings.json` — сначала `.tmp` / `.swap`, потом `os.Rename`. Падение/обесточивание посреди записи больше не обнулит живой файл.
- **Лимит 100 МБ** на загрузку sing-box core (проверка Content-Length до скачивания + кумулятивный счётчик во время). Скомпрометированное или неправильно настроенное зеркало не может залить гигабайты на диск пользователя.
- **Сводка по источникам в парсере:** успех теперь пишет `"Configuration updated: 2/3 source(s) succeeded (1 failed)"`, если какая-то подписка вернула ошибку или ноль нод.
- **Плитка последней ошибки автообновления** в Core Dashboard — показывает фактическое сообщение (с расшифровкой HTTP-кодов: `401 → "токен может быть просрочен"`, `429 → "rate limited — try again later"`). Висит до 24 ч после ошибки, очищается при следующем успехе.
- **Подсказка «(подписки: X ч назад)»** рядом с датой модификации конфига.
- **Маркер `*`** на кнопке Update, если визард сохранил изменения, а парсер ещё не прокатал.

### Добавлено — безопасность

- **Редакция токена Clash API** в debug-логах — больше не утекает целиком при публикации логов в GitHub issues.
- **Блокировка кнопки Ping-all** на время работы теста (не плодит параллельных воркеров на спам-клик).

### Исправлено

- **CI был красный с 2026-04-16:** `internal/locale.TestAllKeysPresent` — не хватало 4 переводов в `ru.json` для jump-outbound share-URI.
- **STUN settings dialog** обрезался на Windows (**#54**) — теперь принудительно ≥ 520 px.

### Изменено — шаблон

- **Список русских и кириллических TLD расширен** в `ru-domains` rule-set: добавлены `.рус / .москва / moscow / tatar / .дети / .сайт / .орг / .ком`.
