# IMPLEMENTATION REPORT: WIREGUARD_URI

## Статус

- [x] Реализовано
- [x] Сборка и тесты проходят (core/config, core/config/subscription; полная сборка с GUI зависит от окружения CGO/OpenGL)

## Дата

2026-03-06

## Изменённые файлы

| Файл | Изменения |
|------|-----------|
| `core/config/subscription/node_parser.go` | IsDirectLink: +wireguard://. ParseNode: ветка wireguard → parseWireGuardURI. parseWireGuardURI, splitAndTrim. Импорт debuglog. |
| `core/config/subscription/node_parser_test.go` | IsDirectLink: кейсы WireGuard. TestParseNode_Wireguard: валидный URI, невалидные (нет publickey/address/allowedips/hostname). |
| `core/config/outbound_generator.go` | OutboundGenerationResult: +EndpointsJSON, EndpointsCount. В цикле по allNodes: wireguard → EndpointsJSON, иначе → OutboundsJSON. GenerateEndpointJSON. |
| `core/config/updater.go` | WriteToConfig: параметр endpointsContent; замена блока @ParserSTART_E/@ParserEND_E при наличии. UpdateConfigFromSubscriptions: передаёт result.EndpointsJSON, проверка «есть что писать» учитывает endpoints. |
| `ui/wizard/models/wizard_model.go` | GeneratedEndpoints []string, OutboundStats.EndpointsCount. NewWizardModel: инициализация GeneratedEndpoints. |
| `ui/wizard/business/parser.go` | Присвоение model.GeneratedEndpoints, model.OutboundStats.EndpointsCount. Условие TemplatePreviewNeedsUpdate учитывает GeneratedEndpoints. |
| `ui/wizard/business/create_config.go` | buildConfigSections: case "endpoints" → buildEndpointsSection. buildEndpointsSection: маркеры @ParserSTART_E/@ParserEND_E, модель GeneratedEndpoints, статические из шаблона. |
| `bin/wizard_template.json` | В config добавлена секция "endpoints": []. |
| `docs/ParserConfig.md` | Поддержка WireGuard в описании, connections, таблице, шагах; подраздел WireGuard (wireguard://); маркеры endpoints; требование sing-box 1.11+. |

## Краткое описание изменений

1. **Парсер**  
   Строки `wireguard://...` распознаются как прямые ссылки. Разбор: userinfo = приватный ключ (декодирование через `url.PathUnescape`), host:port = сервер, обязательные query `publickey`, `address`, `allowedips`; опциональные `mtu`, `keepalive`, `presharedkey`, `listenport`, `name`, `dns`. Для `publickey` и `presharedkey` используется сохранение плюса в query (`queryParamPreservePlus`), чтобы base64 с `+` не превращался в пробел. Строится полная структура sing-box endpoint (type wireguard, tag, name, mtu, address, private_key, listen_port, peers). Поле `listen_port` в endpoint добавляется только при ненулевом значении. Поле `persistent_keepalive_interval` в peer добавляется только при указании `keepalive` (по умолчанию не заполняется — sing-box ставит 0).

2. **Генерация и модель**  
   Ноды с `Scheme == "wireguard"` попадают в `EndpointsJSON`, остальные — в `OutboundsJSON`. В модели визарда добавлены `GeneratedEndpoints` и `OutboundStats.EndpointsCount`. Превью и сохранение собирают секцию `endpoints` из `model.GeneratedEndpoints`.

3. **Сборка конфига**  
   В шаблоне добавлена секция `"endpoints": []` (порядок из JSON даёт ConfigOrder). При сборке вызывается `buildEndpointsSection`: между маркерами @ParserSTART_E и @ParserEND_E вставляются сгенерированные endpoint-строки, при необходимости дополняются статическими из шаблона.

4. **Updater**  
   При обновлении конфига из подписок записываются и outbounds (@ParserSTART/@ParserEND), и endpoints (@ParserSTART_E/@ParserEND_E). Если маркеров endpoints в файле нет, запись outbounds не ломается. Проверка «есть что писать» учитывает оба среза (достаточно одного непустого).

5. **Документация**  
   В ParserConfig.md добавлен WireGuard в список протоколов, пример в connections, подраздел формата wireguard:// и указание, что узлы попадают в секцию endpoints и требуется sing-box 1.11+.

## Замечания / отступления от SPEC

- Только `bin/wizard_template.json` изменён (по уточнению пользователя); остальные шаблоны не трогались.
- `persistent_keepalive_interval`: при отсутствии `keepalive` поле в peer не добавляется (sing-box по умолчанию 0).
- Полная сборка приложения (`go build ./...`) на окружении без CGO/OpenGL может падать из-за Fyne; пакеты `core/config` и `core/config/subscription` собираются и тесты проходят.

## Смежные изменения в рабочих файлах

В тех же файлах (`ui/wizard/business/parser.go`, `ui/wizard/tabs/source_tab.go`, `ui/wizard/outbounds_configurator/*`, `ui/wizard/presentation/*`, `ui/wizard/business/ui_updater.go`) в той же ветке выполнены рефакторинги, не относящиеся к WIREGUARD_URI: модель как единый источник истины (чтение из presenter.Model()), унификация сборки прокси (proxyInput, buildProxiesFromInputs), общий индекс префиксов (1:, 2:, …), UIUpdater с Model(). Полный список — в [docs/release_notes/upcoming.md](../../docs/release_notes/upcoming.md).
