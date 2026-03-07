# PLAN: WIREGUARD_URI

Краткий план реализации поддержки `wireguard://` URI. Детали формата и требований — в [SPEC.md](SPEC.md). Раздел 2.6 SPEC описывает, почему endpoints появляются в стейдже и где их учитывать.

## Компоненты

### 1. Парсер (node_parser, source_loader)

- **IsDirectLink()**: добавить проверку `wireguard://`.
- **ParseNode()**: ветка для схемы `wireguard://` — разбор userinfo@host:port?query, обязательные/опциональные query, построение полной структуры sing-box endpoint в `node.Outbound`, тег из fragment/name. Общий `buildOutbound()` для wireguard не использовать (нет server/server_port).
- **source_loader**: без изменений — `wireguard://` уже будет обрабатываться как прямая ссылка в Source, Connections и по строкам подписки.

### 2. Генерация (стейдж): разделение outbounds / endpoints

- **OutboundGenerationResult**: добавить поле **EndpointsJSON** `[]string` (и при необходимости **EndpointsCount**).
- **GenerateOutboundsFromParserConfig**: при обходе `allNodes` разделять: если `node.Scheme == "wireguard"` — сериализовать endpoint (из `node.Outbound` или отдельная функция) и добавить в срез для endpoints; иначе — как сейчас, `GenerateNodeJSON(node)` в OutboundsJSON. Итог: в `OutboundsJSON` только ноды-не-wireguard + селекторы; в `EndpointsJSON` — только строки JSON для WireGuard-нод.
- **GenerateNodeJSON**: либо расширить для `node.Scheme == "wireguard"` (возвращать JSON endpoint'а в том же формате строки с trailing comma), либо ввести **GenerateEndpointJSON(node)** и вызывать её только для wireguard в цикле генерации. Важно: одна строка на endpoint, формат как у остальных (комментарий + JSON с запятой), чтобы вставлять в массив endpoints.

### 3. Сборка конфига (create_config, шаблон)

- **Шаблон**: в конфиге шаблона (например `wizard_template.json`) и в **ConfigOrder** добавить секцию **"endpoints"** (массив; можно пустой или с комментарием), чтобы в финальном config.json она присутствовала.
- **buildConfigSections**: добавить `case "endpoints":` — вызывать **buildEndpointsSection(model, raw, forPreview, timing)** по аналогии с outbounds (динамические endpoint'ы + при необходимости статические из шаблона; маркеры @ParserSTART/@ParserEND для endpoints опциональны, если решим вставлять только сгенерированное).
- **buildEndpointsSection**: формировать массив JSON из `model.GeneratedEndpoints`; маркеры для вставки динамической части — **@ParserSTART_E** и **@ParserEND_E** (по аналогии с @ParserSTART/@ParserEND для outbounds). Если в шаблоне есть статические элементы — объединить (порядок: сгенерированные, затем статические, или наоборот — зафиксировать в плане/спеце).

### 4. Модель визарда

- **WizardModel**: добавить поле **GeneratedEndpoints** `[]string` (и при необходимости в **OutboundStats** — счётчик endpoint'ов, если нужен для превью/подписей).
- **parser.go** (ParseAndPreview): после генерации присваивать `model.GeneratedEndpoints = result.EndpointsJSON` (и обновлять статистику, если введена).

### 5. Всё, что работает с outbounds

- **filterNodesForSelector / getNodeValue**: уже используют `node.Scheme`; значение `"wireguard"` будет участвовать в фильтрах по scheme — изменений не требуется.
- **Селекторы**: теги WireGuard-нод попадают в `allNodes` и в фильтрацию; в список outbounds селектора можно включать тег WireGuard — в конфиге он будет ссылаться на элемент из `endpoints`. Изменений в GenerateSelectorWithFilteredAddOutbounds не требуется (работа по тегам).
- **Превью (Sources, Rules, Preview, preview_cache)**: работают с общим списком нод; WireGuard-ноды отображаются как остальные. Изменений не требуется, если ноды приходят в кеш с `Scheme: "wireguard"`.
- **Edit Outbound / конфигуратор**: при выборе тегов (default, addOutbounds) список тегов формируется из тех же нод/outbounds; теги endpoint'ов должны входить в доступные варианты — проверить, что источник списка тегов включает все ноды (и outbound-, и endpoint-теги). При необходимости явно объединять теги из OutboundsJSON и EndpointsJSON для UI (если где-то строится список только из «outbound-тегов»).

### 6. Updater (обновление конфига по подпискам)

- **UpdateConfigFromSubscriptions** после генерации получает `result` с `OutboundsJSON` и `EndpointsJSON`. Сейчас в файл пишется только содержимое между @ParserSTART и @ParserEND (outbounds). Нужно также записать сгенерированные endpoints: либо в **WriteToConfig** добавить второй блок (поиск @ParserSTART_E / @ParserEND_E и подстановка `strings.Join(result.EndpointsJSON, "\n")`), либо одна операция чтения/записи файла, которая обновляет и outbounds, и endpoints. Иначе при автообновлении подписок секция endpoints в config.json не будет обновляться.

### 7. Документация и тесты

- **docs/ParserConfig.md**: пример `wireguard://...` в connections/source, WireGuard в списке поддерживаемых форматов; при необходимости кратко упомянуть, что WireGuard попадает в секцию `endpoints` конфига.
- **node_parser_test.go**: тесты `IsDirectLink("wireguard://...")`, `ParseNode` с валидным/невалидным wireguard URI, проверка `Outbound` (type wireguard, peers, address, mtu).
- При появлении **GenerateEndpointJSON** или ветки wireguard в **GenerateNodeJSON** — тест генерации одной WireGuard-ноды в строку JSON (опционально в generator_test или отдельный файл).

## Файлы для изменения

| Файл | Изменения |
|------|-----------|
| `core/config/subscription/node_parser.go` | IsDirectLink: +wireguard. ParseNode: ветка wireguard, построение endpoint в node.Outbound. |
| `core/config/subscription/node_parser_test.go` | Тесты wireguard URI и ParsedNode/Outbound. |
| `core/config/outbound_generator.go` | OutboundGenerationResult: +EndpointsJSON (и счётчик при необходимости). В цикле по allNodes: wireguard → EndpointsJSON, остальные → OutboundsJSON. GenerateEndpointJSON или ветка wireguard в GenerateNodeJSON. |
| `core/config_service.go` | Если интерфейс результата меняется — передавать EndpointsJSON дальше. |
| `ui/wizard/business/parser.go` | Присвоение model.GeneratedEndpoints = result.EndpointsJSON (и статистика). |
| `ui/wizard/business/create_config.go` | buildConfigSections: case "endpoints" → buildEndpointsSection. Реализация buildEndpointsSection(model, raw, ...). |
| `ui/wizard/models/wizard_model.go` | Поле GeneratedEndpoints (и при необходимости OutboundStats.EndpointsCount). |
| `ui/wizard/template/loader.go` (или место, где задаётся ConfigOrder) | Включить "endpoints" в порядок секций; в шаблоне config добавить ключ "endpoints" (пустой массив или с маркерами). |
| `bin/wizard_template.json` (и при необходимости другие шаблоны) | **Задача при реализации:** в блоке config перед `outbounds` добавить секцию `"endpoints": []`. Итог: `"inbounds": [], "endpoints": [], "outbounds": [ ... ]`. Маркеры @ParserSTART_E / @ParserEND_E в шаблоне не хранятся — их выводит код в buildEndpointsSection. Файлы проекта на этапе постановки задачи не правим. |
| `core/config/updater.go` | **UpdateConfigFromSubscriptions**: после генерации писать в config не только outbounds (@ParserSTART/@ParserEND), но и endpoints (маркеры @ParserSTART_E/@ParserEND_E). **WriteToConfig**: расширить для записи блока endpoints или выполнять одну запись файла с обновлением обеих секций. |
| `docs/ParserConfig.md` | Пример wireguard, список форматов, при необходимости — про секцию endpoints. |

## Риски

- Специфика sing-box endpoint: имена полей (`listen_port`, `pre_shared_key`, `persistent_keepalive_interval`) — сверить с [документацией sing-box](https://sing-box.sagernet.org/configuration/endpoint/wireguard/).
- Версия sing-box: endpoints (WireGuard) с 1.11; при использовании старой версии конфиг с `endpoints` может вести себя иначе — в документации пользователя при необходимости указать требование версии.
