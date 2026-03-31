# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

---

## EN

### Internal / Refactoring

### Highlights

### Technical / Internal

- Build scripts `build/build_linux.sh` and `build/test_linux.sh` are stored in git with the executable bit; after clone, run `./build/...` without `chmod +x` on tracked files ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).

---

## RU

### Внутреннее / Рефакторинг

### Основное

### Техническое / Внутреннее

- Скрипты `build/build_linux.sh` и `build/test_linux.sh` в репозитории с флагом исполняемого файла; после клона достаточно `./build/...` без `chmod +x` для отслеживаемых файлов ([issue #49](https://github.com/Leadaxe/singbox-launcher/issues/49)).
