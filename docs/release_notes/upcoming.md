# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

---

## EN

### Internal / Refactoring

(пункты для следующего релиза)

### Highlights

- **Wizard — subscription URL fragment:** If a subscription URL contains a `#fragment` (e.g. `#abvpn`), Apply/Append sets `tag_prefix` from that fragment (sanitized, with a trailing `:` like numeric prefixes) when no `tag_prefix` is already stored for that source.

- **Wizard — UTF-8 labels:** Source/outbound labels are truncated by **Unicode code points** (currently up to **60** visible characters before `...`), not raw bytes, so Cyrillic, emoji flags, and punctuation (e.g. `»`, `❯`) no longer break when the UI shortens long strings. VLESS URI **fragments** are decoded with `PathUnescape` so a literal `+` in the name is not turned into a space. **Preview / server list:** subscription lines and `sanitizeForDisplay` no longer iterate broken UTF-8 (which used to insert U+FFFD); strings are cleaned with `ToValidUTF8` before parse and before Fyne; outbound configurator row text uses the same rune-safe truncation. **Abvpn-style `❯` (U+276F) in tags:** when **reading** subscriptions, `internal/textnorm.NormalizeProxyDisplay` maps `❯` / `»` / `›` to ASCII ` > ` on labels and tags (so generated `config.json` matches what the UI shows). **Servers tab (Clash API):** each `ProxyInfo` keeps the raw API `Name` for requests; `DisplayName` is filled at fetch time with the same normalization for list labels, tray submenu, and status text.

- **VLESS / Trojan subscription links:** Parser and `GenerateNodeJSON` build sing-box [V2Ray transport](https://sing-box.sagernet.org/configuration/shared/v2ray-transport/) from URI query: `ws` (path, headers `Host` — if `host=` is missing, **`sni` is used** for `Host`, e.g. abvpn-style `type=ws&sni=…` only), `http` (`host` as JSON list, path), `grpc` (`service_name`), `xhttp` → `httpupgrade` (only `host` / `path` / `headers` per docs; Xray `mode` is not in the schema). VLESS `security=none` omits TLS; plain TLS and Reality (`pbk`) follow [outbound TLS](https://sing-box.sagernet.org/configuration/shared/tls/#outbound). **REALITY over plain TCP** with no `flow` in the URI gets **`flow: xtls-rprx-vision`** (not applied when transport is `ws`/`grpc`/`http`/`xhttp`). Trojan + WS gets `transport` + `tls`. VMess WS uses the same `host` / `sni` fallback for `Host`. VMess gRPC uses `service_name` from JSON `path`. Wizard preview deduplicates tags like the main parser (`MakeTagUnique`). Query keys are matched case-insensitively where providers use `allowinsecure=0`; multiply-encoded `alpn` is normalized; `fp=QQ` maps to utls `qq`; `tcp`/`raw` with `headerType=http` maps to HTTP transport; `packetEncoding` is copied to outbound `packet_encoding`.

- **VLESS `xtls-rprx-vision-udp443`:** Subscriptions often use Xray’s vision-udp443 flow; sing-box only accepts `xtls-rprx-vision`. The parser already mapped this internally, but generated `config.json` still wrote the original flow and omitted `packet_encoding`. Generation now matches sing-box (vision + `packet_encoding: xudp` when applicable).

- **SOCKS5 in connections:** Parser now supports `socks5://` and `socks://` direct links in Source and Connections (e.g. `socks5://user:pass@proxy.example.com:1080#Office SOCKS5`). Resulting nodes become sing-box outbounds of type `socks` and participate in selectors like other protocols.

- **Linux build:** `build_linux.sh` now checks for required system packages (OpenGL/X11) and prints install commands for Debian/Ubuntu and Fedora. README and new `docs/BUILD_LINUX.md` document dependencies; optional `build/Dockerfile.linux` allows building without installing dev packages locally (see [Issue #40](https://github.com/Leadaxe/singbox-launcher/issues/40)).

- **Wizard — Sources tab:** Scrollable areas (URL field, sources list, server preview, outer tab scroll) reserve a right gutter so the scrollbar does not overlap text or buttons.

- **macOS build script:** `build_darwin.sh` supports `-i` (if the app already exists in `/Applications`, only the executable is updated so `Contents/MacOS/bin/` and logs are kept; otherwise full `.app` copy; then removes the built `.app` from the project directory), `arm64` for a fast Apple Silicon–only build, and `-h` / `--help` (parsed before `go mod tidy`). README documents the options.

- **Wizard template — DNS:** The default `bin/wizard_template.json` DNS section was reworked: `local` resolver, separate UDP servers (e.g. Cloudflare 1.1.1.1 and a Google UDP bootstrap for DoH), Google DoH endpoints use host `dns.google` with `domain_resolver`, and `dns.final` targets the system local resolver. Legacy `bin/config_template.json` and `bin/config_template_macos.json` were removed from the repo. **Recommendation:** delete or reset your saved wizard/parser template in the app data directory so the next run picks up the bundled template and new DNS defaults (otherwise an old copy keeps the previous DNS block).

### Technical / Internal

- **Clash API:** `GET /proxies/{name}/delay` and `PUT /proxies/{group}` now **percent-encode** proxy/group names (spaces, `>`, Unicode, etc.); delay `url` query uses `QueryEscape`. Switch payload uses `json.Marshal` for `name`. Fixes 404 «Resource not found» when pinging tags like `abvpn:… > …`.

- **UI:** `ShowDownloadFailedManual` and `ShowAutoHideInfo` are no longer re-exported from `ui/dialogs.go`; call sites in package `ui` use `internal/dialogs` directly (same behavior).

- **Docs:** `docs/ParserConfig.md` — VLESS/Trojan URI: expanded query parameters and link to `SPECS/023-…/SUBSCRIPTION_PARAMS_REPORT.md` (sing-box field reference); wizard auto `tag_prefix` from subscription URL `#fragment`.

(пункты для следующего релиза)

---

## RU

### Внутреннее / Рефакторинг

(пункты для следующего релиза)

### Основное

- **Визард — фрагмент URL подписки:** если в ссылке на подписку есть `#фрагмент` (например `#abvpn`), при Apply/Append в `tag_prefix` подставляется этот фрагмент (очищенный, с завершающим `:` как у числовых префиксов), если для этого источника ещё не сохранён свой `tag_prefix`.

- **Визард — UTF-8 в подписях:** обрезка длинных подписей источников/строк — по **рунам** (сейчас до **60** символов до `...`), а не по байтам, чтобы не ломать UTF-8 (кириллица, флаги, символы вроде `»` и `❯`). Фрагмент `vless://…#…` декодируется через `PathUnescape`, чтобы `+` в имени не превращался в пробел. **Превью / список серверов:** строки подписки и `sanitizeForDisplay` больше не гоняют по рунам битый UTF-8 (из‑за этого в тег попадал U+FFFD); перед разбором и перед выводом в Fyne применяется `ToValidUTF8`; строки в списке конфигуратора outbounds — та же обрезка по рунам. **Теги с `❯` (U+276F), как у abvpn:** при **чтении** подписки `internal/textnorm.NormalizeProxyDisplay` заменяет `❯`/`»`/`›` на ASCII ` > ` в подписях и тегах (итоговый `config.json` совпадает с тем, что видно в UI). **Вкладка «Серверы» (Clash API):** в `ProxyInfo` сохраняется исходное `Name` для запросов к API; при загрузке списка заполняется `DisplayName` той же нормализацией — список, меню трея и статусные строки показывают его.

- **Ссылки VLESS / Trojan из подписок:** парсер и `GenerateNodeJSON` собирают [V2Ray transport](https://sing-box.sagernet.org/configuration/shared/v2ray-transport/) sing-box из query: для **WS** в заголовок `Host` подставляется **`host` из query**, а если его нет — **`sni`** (как у abvpn: только `type=ws&sni=…`). `http` (поле `host` — список строк), `grpc` (`service_name`), `xhttp` → `httpupgrade`. VLESS: `security=none` без TLS; обычный TLS и Reality (`pbk`) — по [TLS outbound](https://sing-box.sagernet.org/configuration/shared/tls/#outbound). **REALITY по TCP** без `flow` в URI получает **`flow: xtls-rprx-vision`** (не для `ws`/`grpc`/`http`/`xhttp`). Trojan + WS: `transport` и `tls`. VMess WS: тот же fallback `host`/`sni` для `Host`. VMess gRPC: `service_name` из `path` в JSON. Превью в визарде: `MakeTagUnique` как в основном парсере. Ключи query без учёта регистра; `alpn` с многослойным кодированием нормализуется; `fp=QQ` → utls `qq`; `tcp`/`raw` + `headerType=http` → транспорт `http`; `packetEncoding` → `packet_encoding` в outbound.

- **VLESS `xtls-rprx-vision-udp443`:** В подписках часто приходит flow из Xray; sing-box понимает только `xtls-rprx-vision`. Парсер уже переводил значение во внутренней структуре, но в итоговом `config.json` попадал исходный flow без `packet_encoding`. Генерация конфига исправлена (vision + при необходимости `packet_encoding: xudp`).

- **SOCKS5 в connections:** В Source и Connections можно добавлять прямые ссылки `socks5://` и `socks://` (например `socks5://user:pass@proxy.example.com:1080#Office SOCKS5`). Узлы превращаются в outbound типа `socks` и участвуют в селекторах наравне с остальными протоколами.

- **Сборка на Linux:** скрипт `build_linux.sh` проверяет наличие системных пакетов (OpenGL/X11) и выводит команды установки для Debian/Ubuntu и Fedora. В README и в новом `docs/BUILD_LINUX.md` описаны зависимости; добавлен опциональный `build/Dockerfile.linux` для сборки без установки dev-пакетов (см. [Issue #40](https://github.com/Leadaxe/singbox-launcher/issues/40)).

- **Визард — вкладка Sources:** у прокручиваемых блоков (поле URL, список источников, превью серверов, общий скролл вкладки) справа зарезервировано место под полосу прокрутки, чтобы она не наезжала на текст и кнопки.

- **Сборка macOS:** в `build_darwin.sh` флаг `-i` при уже установленном приложении обновляет только исполняемый файл (сохраняются `Contents/MacOS/bin/` и логи), при первой установке копируется весь `.app`, после успеха удаляется собранный `.app` из каталога проекта; режим `arm64`; `-h` / `--help` до `go mod tidy`. В README описаны опции.

- **Шаблон визарда — DNS:** в дефолтном `bin/wizard_template.json` сильно переработана секция DNS: локальный резолвер, отдельные UDP-серверы (в т.ч. Cloudflare 1.1.1.1 и UDP-bootstrap под Google DoH), для Google DoH указан хост `dns.google` с `domain_resolver`, `dns.final` ведёт на системный локальный DNS. Из репозитория убраны устаревшие `bin/config_template.json` и `bin/config_template_macos.json`. **Рекомендация:** удалить или сбросить сохранённый шаблон визарда/парсера в каталоге данных приложения, чтобы при следующем запуске подтянулся встроенный шаблон и новые настройки DNS (иначе останется старая копия с прежней DNS-секцией).

### Техническое / Внутреннее

- **Clash API:** для `GET /proxies/{name}/delay` и `PUT /proxies/{group}` имена прокси/группы **кодируются** (`PathEscape`), параметр `url` в delay — `QueryEscape`; тело переключения — `json.Marshal` для поля `name`. Устраняет 404 при пинге тегов с пробелами и `>` (например abvpn после нормализации).

- **UI:** `ShowDownloadFailedManual` и `ShowAutoHideInfo` больше не реэкспортируются из `ui/dialogs.go`; вызовы в пакете `ui` идут в `internal/dialogs` напрямую (поведение то же).

- **Документация:** `docs/ParserConfig.md` — VLESS/Trojan URI: расширен список query-параметров и ссылка на `SPECS/023-…/SUBSCRIPTION_PARAMS_REPORT.md` (справочник полей sing-box); описан автоматический `tag_prefix` из `#` во вводе визарда.

(пункты для следующего релиза)
