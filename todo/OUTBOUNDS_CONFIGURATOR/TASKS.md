## TASKS: OUTBOUNDS_CONFIGURATOR

### Этап 1. Перестройка вкладок визарда

- [ ] В `wizard.go` (или месте создания табов визарда) выделить из текущего `CreateSourceTab` два таба:
  - `Sources` — работа с источниками и Preview;
  - `Outbounds and ParserConfig` — работа с outbounds и текстовым ParserConfigJSON.
- [ ] Убедиться, что вкладка `Sources` содержит только:
  - многострочное поле `SourceURLEntry` с подсказкой и текущей логикой `ApplyURLToParserConfig` + CheckURL;
  - статус проверки URL;
  - список Sources;
  - read‑only Preview сгенерированных нод/селекторов.
- [ ] Удалить с вкладки `Sources` поле ParserConfigJSON и элементы управления ParserConfig (ParseButton, Config Outbounds, Documentation/ChatGPT).
- [ ] Добавить во визард вторую вкладку `Outbounds and ParserConfig` с контейнером под список outbounds и редактор ParserConfigJSON.

### Этап 2. Вкладка Sources: Sources‑лист и Preview

- [ ] Реализовать список Sources поверх `ParserConfig.ParserConfig.Proxies`:
  - один элемент на каждый `ProxySource`;
  - короткий label (обрезанный URL или `Source N`).
- [ ] Интегрировать tooltip’ы для элементов списка через `fyne-tooltip`:
  - полный URL источника;
  - `tag_prefix`, `tag_postfix`, `tag_mask`;
  - количество и список тегов локальных outbounds этого источника.
- [ ] Разместить read‑only Preview нод/селекторов внизу вкладки `Sources` (re‑use существующего `OutboundsPreview` из `source_tab.go`).
- [ ] Убедиться, что список Sources и tooltip’ы автоматически обновляются при изменении ParserConfig (после `ApplyURLToParserConfig`, после правок в outbounds и ParserConfigJSON).

### Этап 3. Вкладка Outbounds and ParserConfig: список outbounds

- [ ] На базе существующего `ui/wizard/outbounds_configurator/configurator.go`:
  - вынести построение списка outbounds (`collectRows`, `tagsAbove`, Up/Down, Edit/Delete/Add) в отдельную функцию/компонент, возвращающую `fyne.CanvasObject`;
  - использовать этот компонент внутри вкладки `Outbounds and ParserConfig` вместо отдельного окна.
- [ ] Обеспечить, чтобы общий список outbounds строился в порядке:
  - сначала локальные `proxies[i].outbounds` по каждому источнику;
  - затем глобальные `ParserConfig.outbounds`.
- [ ] Для каждой строки списка отрисовать:
  - иконки Up/Down слева (только внутри своего scope);
  - текст `tag (type) — SourceLabel` (`Global` или label источника);
  - справа кнопки Edit и Delete.
- [ ] Реализовать поведение кнопки Up:
  - доступна, если outbound не первый в своём scope;
  - меняет местами текущий элемент с предыдущим в том же слайсе (переиспользовать `moveOutboundUp`).
- [ ] Реализовать поведение кнопки Down:
  - доступна, если outbound не последний в своём scope;
  - меняет местами текущий элемент со следующим в том же слайсе (переиспользовать `moveOutboundDown`).
- [ ] Добавить кнопку Add, открывающую диалог создания outbound.

### Этап 4. Диалог Edit/Add Outbound

- [ ] Переиспользовать существующий диалог `ShowEditDialog` из `edit_dialog.go`, вызвая его из новой вкладки `Outbounds and ParserConfig` (parent = окно визарда).
- [ ] Добавить/проверить поле Scope:
  - опции For all (глобальный outbound) и For source: `<SourceLabel>` (локальный).
- [ ] Добавить/проверить поля Tag, Type (manual/auto), Comment.
- [ ] Реализовать/уточнить блок Filters:
  - фиксированный ключ `tag` (лейбл, нередактируемый);
  - редактируемое значение‑паттерн (в том числе `!/regex/i` и т.п.).
- [ ] Реализовать/уточнить блок Preferred default:
  - фиксированный ключ `tag`;
  - строка‑паттерн выбора узла по умолчанию.
- [ ] Реализовать/уточнить блок AddOutbounds:
  - чекбоксы `direct-out`, `reject`;
  - чекбоксы по тегам других outbounds, находящихся **выше** в общем списке.
- [ ] Оформить диалог с вертикальным скроллом и отступом справа под полосу прокрутки, чтобы скролл не перекрывал поля (как в текущем `edit_dialog.go`).

### Этап 5. Синхронизация ParserConfig

- [ ] Убедиться, что изменения в `SourceURLEntry` обрабатываются через `ApplyURLToParserConfig` и обновляют структуру `config.ParserConfig` в модели визарда.
- [ ] После успешного применения URL:
  - нормализовать ParserConfig;
  - пересериализовать в `ParserConfigJSON`;
  - обновить редактор ParserConfigJSON во второй вкладке и списки Sources/Outbounds.
- [ ] Реализовать операции Edit/Add/Delete/Up/Down для списка outbounds так, чтобы они:
  - изменяли только структуру ParserConfig (локальные и глобальные outbounds) в модели визарда;
  - после этого вызывали нормализацию и пересериализацию в `ParserConfigJSON`;
  - обновляли UI (список outbounds, Preview нод/селекторов).
- [ ] Настроить обработку ручного редактирования ParserConfigJSON во второй вкладке:
  - парсинг текста при потере фокуса / по debounce / при переключении вкладки;
  - при успешном парсинге — замену структуры `config.ParserConfig` и пересчёт списков Sources/Outbounds;
  - при ошибке — показ `dialogs.ShowError` и откат текста к последнему валидному состоянию.

### Этап 6. Интеграция с Rules/Preview и финальная проверка

- [ ] Привязать запуск Parse/Preview к изменениям ParserConfig и/или переходу на вкладки Rules/Preview (используя существующие механизмы `TriggerParseForPreview` / `UpdateTemplatePreviewAsync` вместо кнопки Parse).
- [ ] Убедиться, что вкладки Rules/Preview используют актуальный ParserConfig и корректно отражают изменения из обеих новых вкладок.
- [ ] Постепенно убрать отдельное окно Config Outbounds:
  - на первом шаге — использовать общую реализацию списка/диалога и для вкладки, и для окна;
  - после стабилизации новой вкладки — удалить кнопку `Config Outbounds` из `source_tab.go` и функцию `Show` из `outbounds_configurator`.
- [ ] Пройтись по критериям приёмки SPEC и проверить, что все пункты выполняются.
