# SPEC: OUTBOUNDS_CONFIGURATOR — новая вкладочная структура визарда

## Проблема

Текущая вкладка **"Sources and ParserConfig"** в визарде совмещает сразу три разных задачи:
- ввод URL подписок и прямых ссылок (источники proxies),
- редактирование целого ParserConfig JSON,
- управление генерацией outbounds (Parse/Preview).

Это создаёт перегруженный UI, затрудняет понимание связей между источниками и outbounds, а также дублирует логику в отдельном окне **Config Outbounds**.

## Цель

Перестроить визард на две вкладки и встроенный конфигуратор outbounds так, чтобы:
- первая вкладка занималась только **источниками (Sources)** и быстрым просмотром результата парсинга;
- вторая вкладка занималась **outbounds и ParserConfig**;
- ParserConfig всегда был синхронизирован с действиями пользователя **без отдельной кнопки Parse**;
- пользователь не выходил из визарда в отдельное окно, а настраивал всё внутри одной многовкладочной формы.

## Общее поведение для пользователя

1. **Вкладка Sources**
   - Пользователь думает в терминах «подписки/ссылки → какие ноды получились».
   - Он вводит URL подписок и прямые ссылки в многострочное поле и видит:
     - подсказку по форматам URL и схемам;
     - статус проверки URL;
     - список источников (по одному на каждый `ProxySource`) с коротким label.
   - При наведении на элемент списка показывается tooltip с:
     - полным URL источника;
     - текущими `tag_prefix` / `tag_postfix` / `tag_mask` (если заданы);
     - количеством и списком тегов локальных outbounds этого источника.
   - Внизу вкладки есть только read‑only Preview сгенерированных нод/селекторов со скроллом.
   - На вкладке нет текста ParserConfig JSON и нет кнопки Parse.

2. **Вкладка Outbounds and ParserConfig**
   - Пользователь управляет всеми outbounds в одном общем списке:
     - сначала локальные outbounds по каждому источнику;
     - затем глобальные outbounds из ParserConfig.
   - В каждой строке списка видны:
     - стрелки Up/Down для изменения порядка в пределах своего scope (локальные или глобальные outbounds);
     - tag и type;
     - метка источника (Global или label source);
     - кнопки Edit и Delete.
   - Кнопка Add открывает диалог создания/редактирования outbound.
   - В нижней/правой части вкладки расположен многострочный редактор ParserConfigJSON с кнопками Documentation и ChatGPT; отдельной кнопки Parse нет.

3. **Синхронизация ParserConfig**
   - Источник правды — структура `config.ParserConfig` в модели визарда.
   - Любые изменения:
     - текста в SourceURLEntry на вкладке Sources;
     - состава и порядка outbounds в конфигураторе;
     - ручного текста ParserConfigJSON
     приводят к обновлению структуры `config.ParserConfig`, её нормализации и пересериализации в ParserConfigJSON.
   - Списки Sources и Outbounds всегда пересчитываются из текущей структуры ParserConfig.

4. **Диалог Edit/Add Outbound**
   - Открывается из вкладки Outbounds and ParserConfig поверх основного окна визарда.
   - Позволяет задать:
     - scope (For all или For source: \<label\>);
     - tag, type (manual/auto), comment;
     - filters и preferredDefault с фиксированным ключом `tag`;
     - дополнительные outbounds (direct-out, reject и другие теги, расположенные выше в общем списке).
   - Весь контент диалога лежит внутри вертикального скролла с отступом под полосу прокрутки.

5. **Остальные вкладки визарда**
   - Вкладки Rules/Preview продолжают работать как сейчас, но всегда читают актуальный ParserConfig.
   - Parse/Preview триггерится автоматически на основе изменений ParserConfig и/или перехода на соответствующие вкладки.
   - Отдельного окна Config Outbounds больше нет — всё управление outbounds живёт во вкладке Outbounds and ParserConfig.

## Требования

### 1. Вкладки визарда

1. **Вкладка 1: Sources**
   - Содержимое:
     - многострочное поле SourceURLEntry (URL подписок + прямые ссылки), как сейчас;
     - подсказка под полем (поддерживаемые схемы, одна ссылка на строку и т.п.);
     - статус проверки URL (CheckURL, индикатор прогресса, текст статуса).
  - **Список Sources**:
     - компонент вида widget.List (или аналогичные контейнер + лейблы);
     - каждый элемент соответствует одному config.ProxySource из ParserConfig.ParserConfig.Proxies;
     - отображаемый текст: короткий label (обрезанный URL или Source 1, Source 2, ...), чтобы влезать по ширине;
     - при наведении на элемент показывается **tooltip** (через уже используемый в проекте пакет `github.com/dweymouth/fyne-tooltip`), содержащий:
       - полный source URL;
       - текущие `tag_prefix`, `tag_postfix`, `tag_mask` (если заданы);
       - количество локальных outbounds для этого источника и список их тегов.
   - **Preview нод / outbounds**:
     - нижний блок показывает текстовое preview (read‑only) сгенерированных узлов и селекторов, аналогично текущему OutboundsPreview;
     - поддерживается вертикальный/горизонтальный скролл.
   - На вкладке **нет** поля ParserConfig JSON и **нет** кнопки Parse.

2. **Вкладка 2: Outbounds and ParserConfig**
  - Верх/лево: **список outbounds** (конфигуратор):
     - порядок: сначала все локальные outbounds по каждому ProxySource (proxies[i].outbounds), затем глобальные (ParserConfig.outbounds);
     - каждая строка содержит:
       - иконки **↑** и **↓** (Fyne `theme.MoveUpIcon` / `theme.MoveDownIcon`) слева от текста;
       - текст `tag (type) — SourceLabel`, где SourceLabel = Global или короткий label источника (как в списке Sources);
       - справа — кнопки **Edit** и **Delete**;
     - кнопка **Up**:
       - доступна только если outbound не первый в своём scope;
       - меняет местами текущий outbound с предыдущим в том же слайсе:
         - локальные: внутри proxies[i].outbounds;
         - глобальные: внутри ParserConfig.outbounds;
     - кнопка **Down**:
       - доступна только если outbound не последний в своём scope;
       - меняет местами текущий outbound со следующим в том же слайсе;
     - кнопка **Add** под/над списком открывает диалог создания outbound (см. ниже);
     - после любых изменений список пересчитывается из актуального ParserConfig.
   - Низ/право: **редактор ParserConfig JSON**:
     - многострочное поле, содержащее сериализованный config.ParserConfig (поле ParserConfigJSON модели);
     - кнопки **Documentation** и **ChatGPT** (как сейчас);
     - кнопки **Parse** на этой вкладке **нет** — обновление делается автоматически.

### 2. Синхронизация ParserConfig

3. Единый источник правды — структура config.ParserConfig в модели визарда (model.ParserConfig).
4. ParserConfigJSON всегда является сериализацией этой структуры (через существующую нормализацию), а не независимым текстом.
5. **Обновление из вкладки Sources**:
   - изменение текста в SourceURLEntry по‑прежнему проходит через ApplyURLToParserConfig (debounce, CheckURL — без изменений по смыслу);
   - ApplyURLToParserConfig обновляет model.ParserConfig (struct);
   - после успешного применения:
     - ParserConfig нормализуется и сериализуется в ParserConfigJSON;
     - редактор ParserConfig на вкладке Outbounds and ParserConfig получает новое значение (через SyncModelToGUI/UpdateParserConfig);
     - список Sources и outbounds пересчитывается из обновлённого ParserConfig.
6. **Обновление из вкладки Outbounds and ParserConfig (через UI)**:
   - Edit/Add/Delete/Up/Down в списке outbounds модифицируют model.ParserConfig (структуру: локальные и глобальные outbounds);
   - после каждой операции:
     - выполняется нормализация ParserConfig (version, reload, last_updated — по текущей логике);
     - структура сериализуется обратно в ParserConfigJSON;
     - редактор ParserConfig обновляется; Preview/Rules/Preview‑tab видят новые данные.
7. **Ручное редактирование ParserConfig JSON**:
   - пользователь может править текст в редакторе на вкладке Outbounds and ParserConfig;
   - при потере фокуса / по debounce / при переключении вкладки:
     - JSON валидируется и парсится в config.ParserConfig;
     - при успехе model.ParserConfig заменяется новой структурой, списки Sources/Outbounds пересчитываются;
     - при ошибке отображается понятное сообщение через dialogs.ShowError, редактор откатывается к последнему валидному ParserConfigJSON.

### 3. Диалог Edit/Add Outbound

8. Диалог редактирования/создания outbound открывается из вкладки Outbounds and ParserConfig и работает поверх неё (как сейчас окно configurator, но логически относится к вкладке).
 9. Поля диалога:
   - **Scope**: Select For all (глобальный outbound) или For source: <SourceLabel> (локальный для конкретного источника);
   - **Tag**: текстовое поле;
   - **Type**: Select `manual (selector)` / `auto (urltest)`;
   - **Comment**: текстовое поле (опционально);
   - **Filters**:
     - ключ зафиксирован как `tag` (лейбл, нередактируемый);
     - значение — строка‑паттерн (в т.ч. с отрицанием через !/regex/i и т.п.);
   - **Preferred default (preferredDefault)**:
     - ключ зафиксирован как `tag` (лейбл);
     - значение — строка‑паттерн для выбора узла по умолчанию (например, /🇳🇱/i);
   - **AddOutbounds**:
     - чекбоксы `direct-out`, `reject`;
     - чекбоксы по тегам других outbounds, которые находятся **выше** в текущем списке (локальные + глобальные) — чтобы зависимости были направлены только вниз.
10. Визуальные требования к диалогу:
    - контент вложен в вертикальный скролл, чтобы влезать по высоте окна;
    - справа внутри скролла зарезервирован прозрачный отступ (gap), чтобы полоса прокрутки не заезжала на поля ввода;
    - ширина формы фиксирована/минимальная, чтобы иконки и текст не ломались.

### 4. Список Sources и tooltip’ы

11. Список Sources на первой вкладке реализуется поверх уже существующей модели ParserConfig.ParserConfig.Proxies.
12. Для tooltip’ов используются компоненты из github.com/dweymouth/fyne-tooltip (см. уже реализованный PING_ERROR_TOOLTIP и раздел в ARCHITECTURE.md):
    - окно визарда уже обёрнуто в AddWindowToolTipLayer(content, canvas);
    - элементы списка/лейблы создаются как tooltip‑совместимые виджеты и получают SetToolTip(...) с текстом, описанным выше.
13. При изменении ParserConfig (URL, outbounds) tooltip‑данные должны обновляться автоматически при следующей отрисовке списка.

### 5. Поведение Preview и Rules

14. Логика генерации outbounds и Preview (ParseAndPreview, Rules‑tab, Preview‑tab) не меняется по сути, но опирается на обновлённый ParserConfig и ParserConfigJSON:
    - если раньше генерация привязывалась к кнопке Parse, теперь запуск Parse/Preview должен быть привязан к событиям изменения ParserConfig (URL, правки в Outbounds/JSON) и/или к переходу на Rules/Preview вкладки (как уже реализовано через TriggerParseForPreview/UpdateTemplatePreviewAsync).

## Критерии приёмки

- [ ] Первая вкладка визарда называется **Sources**, не содержит поля ParserConfig JSON и кнопки Parse.
- [ ] На вкладке Sources есть список источников (по ProxySource) и tooltip при наведении показывает полный URL и параметры (`tag_prefix`, `tag_postfix`, `tag_mask`, локальные outbounds).
- [ ] Вторая вкладка называется **Outbounds and ParserConfig** и содержит:
  - список outbounds (локальные по источникам → глобальные) с иконками Up/Down, кнопками Edit/Delete и кнопкой Add;
  - редактор ParserConfig JSON с кнопками Documentation и ChatGPT.
- [ ] Кнопки Up/Down перемещают outbound только внутри своего scope (локальные/глобальные) и корректно обновляют порядок в ParserConfig.
- [ ] Отдельного окна Config Outbounds больше нет; весь функционал конфигуратора реализован во вкладке Outbounds and ParserConfig.
- [ ] Любые изменения URL/links на вкладке Sources и любые операции с outbounds на вкладке Outbounds and ParserConfig приводят к обновлению model.ParserConfig и согласованного ParserConfigJSON **без ручного Parse**.
- [ ] Ручное редактирование ParserConfig JSON при валидном содержимом обновляет структуру ParserConfig и списки Sources/Outbounds; при ошибке JSON пользователь видит понятное сообщение, и текст откатывается к последнему корректному состоянию.
- [ ] Диалог Edit/Add соответствует полям и UX, описанным в разделе 3 (фиксированные ключи tag для filters и preferredDefault, addOutbounds только к тегам выше).
- [ ] Поведение Rules‑/Preview‑вкладок не ломается: они продолжают использовать актуальный ParserConfig и корректно отражать изменения.
- [ ] Все новые сообщения в UI на английском; логирование и обработка ошибок соответствуют constitution.md и IMPLEMENTATION_PROMPT.md (debuglog, dialogs.ShowError, отсутствие немотивированных _ = err).
