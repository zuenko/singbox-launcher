# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

## EN
### Highlights
- **Xray sidecar for XHTTP transport**: sing-box does not support native `xhttp`. The launcher now automatically launches an Xray-core sidecar process for VLESS+XHTTP outbounds. sing-box routes traffic through a local SOCKS5 proxy handled by Xray, giving full XHTTP support including `mode` and `extra` fields. If Xray core is not installed, XHTTP outbounds gracefully fall back to `httpupgrade`.
- **Xray core management in Core Dashboard**: new block showing Xray installation status and version, plus one-click download from GitHub releases (XTLS/Xray-core).
- **Auto-download Xray**: if xray.exe is missing and the subscription cache contains XHTTP nodes, the launcher automatically downloads Xray core in the background on startup.

### Technical / Internal
- Added `core/xray/` package: config builder (sing-box outbound → Xray JSON), process manager, port registry.
- Build pipeline (`core/build`) intercepts `transport.type == "xhttp"` outbounds and transforms them to `type: "socks"` pointing to the sidecar, or falls back to `httpupgrade`.
- ProcessService lifecycle extended: start/stop/restart Xray sidecar alongside sing-box.
- Proxy switch callback triggers sidecar restart when the active outbound changes to/from XHTTP.

## RU
### Основное
- **Xray sidecar для XHTTP-транспорта**: sing-box не поддерживает нативный `xhttp`. Лauncher теперь автоматически запускает Xray-core sidecar для VLESS+XHTTP-нод. sing-box отправляет трафик через локальный SOCKS5-прокси, который обрабатывает Xray — полная поддержка XHTTP с `mode` и `extra`. Если Xray core не установлен, XHTTP-ноды автоматически откатываются на `httpupgrade`.
- **Управление Xray core в Core Dashboard**: новый блок со статусом установки и версией Xray, а также кнопка скачивания из GitHub releases (XTLS/Xray-core).
- **Автоскачка Xray**: если xray.exe отсутствует и в кэше есть XHTTP-ноды, launcher автоматически скачивает Xray core в фоне при старте.

### Техническое / Внутреннее
- Новый пакет `core/xray/`: конвертер конфига (sing-box outbound → Xray JSON), менеджер процесса, реестр портов.
- Сборка конфига (`core/build`) перехватывает outbounds с `transport.type == "xhttp"` и заменяет их на `type: "socks"` → sidecar, либо fallback на `httpupgrade`.
- ProcessService расширен: старт/стоп/рестарт Xray sidecar вместе с sing-box.
- Callback переключения прокси перезапускает sidecar при смене активной ноды на XHTTP и обратно.
