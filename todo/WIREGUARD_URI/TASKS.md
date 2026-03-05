# TASKS: WIREGUARD_URI

Чеклист по [SPEC.md](SPEC.md) и [PLAN.md](PLAN.md). Endpoints появляются в стейдже; всё, что работает с outbounds, должно учитывать endpoints (SPEC 2.6, PLAN).

## Этап 1: Парсер

- [ ] В `IsDirectLink()` добавить распознавание `wireguard://`.
- [ ] В `ParseNode()` добавить ветку для `wireguard://`: парсинг userinfo (private key), host, port, query.
- [ ] Извлечь и провалидировать обязательные query: `publickey`, `address`, `allowedips`.
- [ ] Извлечь опциональные: `mtu`, `keepalive`, `presharedkey`, `listenport`, `name`, `dns`.
- [ ] Построить `ParsedNode`: Scheme=`wireguard`, Tag (из fragment или name), Server, Port, Outbound = полная структура sing-box wireguard endpoint (см. SPEC 3.5).
- [ ] Применить дефолты: mtu=1420, listenport=0; не добавлять в peer поля pre_shared_key / persistent_keepalive_interval, если не заданы (или 0 по рекомендации).
- [ ] Обработка ошибок и логирование (debuglog) при невалидном URI.

## Этап 2: Стейдж — разделение outbounds / endpoints

- [ ] В `OutboundGenerationResult` добавить поле `EndpointsJSON []string` (и при необходимости счётчик).
- [ ] В `GenerateOutboundsFromParserConfig`: при обходе `allNodes` разделять — `node.Scheme == "wireguard"` → в `EndpointsJSON`, иначе → `GenerateNodeJSON(node)` в `OutboundsJSON` как сейчас.
- [ ] Реализовать сериализацию одного WireGuard-endpoint в JSON-строку (функция типа `GenerateEndpointJSON(node)` или ветка в `GenerateNodeJSON` для wireguard); формат строки с trailing comma для вставки в массив.
- [ ] Убедиться, что селекторы по-прежнему могут ссылаться на теги WireGuard-нод (теги входят в outbounds списки селекторов, физически ноды — в endpoints).

## Этап 3: Модель и парсер-координатор

- [ ] В `WizardModel` добавить `GeneratedEndpoints []string` (и при необходимости счётчик в `OutboundStats`).
- [ ] В `parser.go` (ParseAndPreview): присвоить `model.GeneratedEndpoints = result.EndpointsJSON` после генерации.

## Этап 4: Updater — обновление endpoints при «обновить из подписок»

- [ ] В `UpdateConfigFromSubscriptions`: использовать не только `result.OutboundsJSON`, но и `result.EndpointsJSON`; записывать endpoints в config (между @ParserSTART_E и @ParserEND_E).
- [ ] Расширить `WriteToConfig` (или добавить запись в том же проходе): искать в файле маркеры @ParserSTART_E / @ParserEND_E и подставлять содержимое `EndpointsJSON`. Если маркеров нет (старый config), не ломать запись outbounds; при необходимости документировать, что для обновления endpoints нужен шаблон с секцией endpoints и маркерами.

## Этап 5: Сборка конфига — секция endpoints

- [ ] В шаблоне конфига (например `bin/wizard_template.json`) добавить секцию `"endpoints": []`; в `ConfigOrder` / загрузке шаблона включить `"endpoints"`.
- [ ] В `buildConfigSections` добавить `case "endpoints":` → вызов `buildEndpointsSection(model, raw, forPreview, timing)`.
- [ ] Реализовать `buildEndpointsSection`: массив JSON из `model.GeneratedEndpoints` (и при необходимости статические из шаблона).

## Этап 6: UI и места, работающие с outbounds

- [ ] Проверить, что список тегов для селекторов / Edit Outbound / default включает теги WireGuard-endpoint'ов (источник тегов — все ноды или объединение OutboundsJSON + EndpointsJSON по тегам).
- [ ] Превью (Sources, Rules, Preview, кеш): убедиться, что WireGuard-ноды отображаются в списках узлов (достаточно Scheme и тега в ParsedNode).

## Этап 7: Тесты

- [ ] Тест `IsDirectLink` для `wireguard://...` (true) и без префикса (false).
- [ ] Тест `ParseNode` с валидным wireguard URI: проверка Tag, Scheme, Outbound.type=wireguard, peers[0].address/port/public_key, address, mtu.
- [ ] Тест с невалидным URI (нет publickey / address / allowedips) — ошибка, нода не создаётся.
- [ ] При появлении генерации endpoints — тест, что WireGuard-нода даёт строку в EndpointsJSON, не в OutboundsJSON (опционально).

## Этап 8: Документация

- [ ] В `docs/ParserConfig.md` в примерах connections добавить строку с `wireguard://`.
- [ ] В списке поддерживаемых протоколов указать WireGuard (wireguard://).
- [ ] При необходимости кратко указать, что WireGuard попадает в секцию `endpoints` config.json.

## Этап 9: Отчёт

- [ ] Заполнить IMPLEMENTATION_REPORT.md после реализации.
