# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

---

## EN

### Internal / Refactoring

(пункты для следующего релиза)

### Highlights

- **SOCKS5 in connections:** Parser now supports `socks5://` and `socks://` direct links in Source and Connections (e.g. `socks5://user:pass@proxy.example.com:1080#Office SOCKS5`). Resulting nodes become sing-box outbounds of type `socks` and participate in selectors like other protocols.

- **Linux build:** `build_linux.sh` now checks for required system packages (OpenGL/X11) and prints install commands for Debian/Ubuntu and Fedora. README and new `docs/BUILD_LINUX.md` document dependencies; optional `build/Dockerfile.linux` allows building without installing dev packages locally (see [Issue #40](https://github.com/Leadaxe/singbox-launcher/issues/40)).

### Technical / Internal

(пункты для следующего релиза)

---

## RU

### Внутреннее / Рефакторинг

(пункты для следующего релиза)

### Основное

- **SOCKS5 в connections:** В Source и Connections можно добавлять прямые ссылки `socks5://` и `socks://` (например `socks5://user:pass@proxy.example.com:1080#Office SOCKS5`). Узлы превращаются в outbound типа `socks` и участвуют в селекторах наравне с остальными протоколами.

- **Сборка на Linux:** скрипт `build_linux.sh` проверяет наличие системных пакетов (OpenGL/X11) и выводит команды установки для Debian/Ubuntu и Fedora. В README и в новом `docs/BUILD_LINUX.md` описаны зависимости; добавлен опциональный `build/Dockerfile.linux` для сборки без установки dev-пакетов (см. [Issue #40](https://github.com/Leadaxe/singbox-launcher/issues/40)).

### Техническое / Внутреннее

(пункты для следующего релиза)
