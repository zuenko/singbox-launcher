# Upcoming release — черновик

Сюда складываем пункты, которые войдут в следующий релиз. Перед релизом переносим в `X-Y-Z.md` и очищаем этот файл.

**Не добавлять** сюда мелкие правки **только UI** (порядок виджетов, выравнивание, стиль кнопок без смены действия и т.п.). Писать **новое поведение**: данные, форматы, сохранение, заметные для пользователя возможности.

---

## EN

### Internal / Refactoring

### Highlights

- Wizard DNS scalars (**strategy**, **independent cache**, **final**, **default domain resolver**) are stored in **`state.vars`** as hidden template variables **`dns_*`** with **`@dns_*`** placeholders in **`config`**; **`dns_options`** in state keeps **servers** and **rules** only (legacy scalar keys migrate on load). See **docs/WIZARD_STATE.md** and **SUB_SPEC_DNS_TAB_VARS**.

### Technical / Internal

- Wizard **`state.json`** root format version **4** on save (reads **2–4**). **4** aligns with the template-**`vars`** era (**`state.vars`**, **`@name`**, **`if`**/**`if_or`** on **params**, **SPECS/032**); **3** remains valid for older snapshots (rules library). See **docs/WIZARD_STATE.md**.

---

## RU

### Внутреннее / Рефакторинг

### Основное

- Скаляры вкладки DNS (strategy, кеш, final, резолвер по умолчанию) хранятся в **`state.vars`** как скрытые **`dns_*`**; в **`dns_options`** остаются **servers** и **rules** (старые ключи при загрузке мигрируют в **`vars`**). См. **docs/WIZARD_STATE.md** и **SUB_SPEC_DNS_TAB_VARS**.

### Техническое / Внутреннее

- Корневой **`version`** **`state.json`** при сохранении — **4** (чтение **2–4**). **4** — линия с переменными шаблона (**`vars`** в state, **`@…`** и **`if`**/**`if_or`** в шаблоне, **032**); **3** — по-прежнему для старых снимков (rules library). См. **docs/WIZARD_STATE.md**.
