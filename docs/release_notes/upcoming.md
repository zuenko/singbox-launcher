# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

---

## EN

### Internal / Refactoring

### Highlights

- **Linux:** If `sing-box` is on `PATH` (e.g. installed from your distro package), the launcher uses it automatically; otherwise it uses `bin/sing-box` next to the launcher. **Core → Download** still installs into local `bin/` only ([issue #48](https://github.com/Leadaxe/singbox-launcher/issues/48)).

### Technical / Internal

- Build scripts `build/build_linux.sh` and `build/test_linux.sh` are stored in git with the executable bit; after clone, run `./build/...` without `chmod +x` on tracked files ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).

---

## RU

### Внутреннее / Рефакторинг

### Основное

- **Linux:** если `sing-box` есть в `PATH` (например, из пакета дистрибутива), лаунчер использует его; иначе — `bin/sing-box` рядом с лаунчером. Кнопка **Core → Download** по-прежнему кладёт бинарник только в локальный `bin/` ([issue #48](https://github.com/Leadaxe/singbox-launcher/issues/48)).

### Техническое / Внутреннее

- Скрипты `build/build_linux.sh` и `build/test_linux.sh` в репозитории с флагом исполняемого файла; после клона достаточно `./build/...` без `chmod +x` для отслеживаемых файлов ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).
