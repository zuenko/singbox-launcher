# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

---

## EN

### Internal / Refactoring

### Highlights

- **Wizard — Settings:** `wizard_template.json` declares configurable **`vars`**; the wizard **Settings** tab shows them, saves values in wizard state, and they substitute **`@name`** placeholders in the generated config.

- **Win7 wizard template:** Loading `wizard_template.json` no longer fails on the Win7 x86 build: `inbounds` is a JSON array, so the old `inbounds.stack` param could not be applied after the TUN block. Win7 now uses a dedicated full `inbounds` replace with `stack: gvisor` (after the shared `windows`/`linux` entry).

- **Linux:** If `sing-box` is on `PATH` (e.g. installed from your distro package), the launcher uses it automatically; otherwise it uses `bin/sing-box` next to the launcher. **Core → Download** still installs into local `bin/` only ([issue #48](https://github.com/Leadaxe/singbox-launcher/issues/48)).

### Technical / Internal

- **Hysteria2 ports from subscriptions:** `mport` / `ports` now follow the official Hysteria 2 list format (comma-separated ports and `start-end` ranges). Multi-port in the URI authority (e.g. `host:443,20000-30000`) is recovered when `net/url` cannot parse it. Bare single ports map to `low:high` for sing-box `server_ports`.

- Build scripts `build/build_linux.sh` and `build/test_linux.sh` are stored in git with the executable bit; after clone, run `./build/...` without `chmod +x` on tracked files ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).

---

## RU

### Внутреннее / Рефакторинг

### Основное

- **Визард — «Настройки»:** В шаблоне (`wizard_template.json`) объявляются пользовательские **`vars`**; лаунчер выводит их на вкладку **«Настройки»**, сохраняет в состоянии визарда и подставляет в собираемый конфиг по плейсхолдерам **`@name`**.

- **Визард Win7:** Загрузка `wizard_template.json` на сборке Win7 x86 больше не падает: у `inbounds` массив, поэтому параметр `inbounds.stack` после блока TUN давал ошибку парсера. Для Win7 — отдельная полная замена `inbounds` со `stack: gvisor` (после общего параметра для `windows`/`linux`).

- **Linux:** если `sing-box` есть в `PATH` (например, из пакета дистрибутива), лаунчер использует его; иначе — `bin/sing-box` рядом с лаунчером. Кнопка **Core → Download** по-прежнему кладёт бинарник только в локальный `bin/` ([issue #48](https://github.com/Leadaxe/singbox-launcher/issues/48)).

### Техническое / Внутреннее

- **Hysteria2 порты из подписок:** Параметры `mport` / `ports` разбираются в официальном формате Hysteria 2 (список через запятую: порты и диапазоны `начало-конец`). Multi-port в authority URI (например `host:443,20000-30000`) восстанавливается, если стандартный парсер URL не принимает строку. Одиночный порт даёт диапазон `n:n` для sing-box `server_ports`.

- Скрипты `build/build_linux.sh` и `build/test_linux.sh` в репозитории с флагом исполняемого файла; после клона достаточно `./build/...` без `chmod +x` для отслеживаемых файлов ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).
