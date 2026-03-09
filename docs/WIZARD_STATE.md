# Wizard state (state.json)

Формат файла состояния визарда конфигурации и логика загрузки/сохранения.

## Назначение

Файл `state.json` (и именованные состояния `<id>.json`) хранит полное состояние визарда: выбранные источники прокси, outbounds, правила маршрутизации (в т.ч. пользовательские), параметры конфигурации. При открытии визарда состояние загружается из текущего файла; при сохранении — записывается обратно.

## Версия формата

- **version**: целое число (текущая версия — `2`). Используется для миграций при загрузке старых файлов.

## Структура JSON

Корневой объект содержит:

| Поле | Тип | Описание |
|------|-----|----------|
| `version` | int | Версия формата (обязательное) |
| `id` | string | Идентификатор состояния (для именованных состояний; опционально для state.json) |
| `comment` | string | Комментарий (опционально) |
| `created_at` | string | RFC3339 (обязательное) |
| `updated_at` | string | RFC3339 (обязательное) |
| `parser_config` | object | Конфигурация парсера (proxies, outbounds, parser) |
| `config_params` | array | Параметры конфигурации (route.final и др.) |
| `selectable_rule_states` | array | Состояния правил из шаблона (label, enabled, selected_outbound) |
| `custom_rules` | array | Пользовательские правила (полная структура) |

## custom_rules (PersistedCustomRule)

Каждый элемент — объект с полями:

| Поле | Тип | Описание |
|------|-----|----------|
| `label` | string | Название правила |
| `type` | string | Тип: только `ips`, `urls`, `processes`, `srs`, `raw` |
| `enabled` | bool | Включено ли правило |
| `selected_outbound` | string | Выбранный outbound |
| `description` | string | Описание (опционально) |
| `rule` | object | JSON объекта правила маршрутизации (ip_cidr, domain, rule_set и т.д.) |
| `default_outbound` | string | Outbound по умолчанию |
| `has_outbound` | bool | Есть ли outbound в правиле |
| `params` | object | Состояние UI по типу (опционально; в конфиг не попадает) |
| `rule_set` | array | Определения rule-set'ов для типа `srs` (опционально) |

### type — константы

В state и в коде используются только значения: `ips`, `urls`, `processes`, `srs`, `raw`. При загрузке, если `type` отсутствует или имеет старый формат (например `"Domains/URLs"`), тип выводится из содержимого `rule` функцией **DetermineRuleType(rule)**. При сохранении всегда записываются только эти константы.

### params

Объект для восстановления состояния интерфейса по типу правила:

- **processes:** `match_by_path` (bool), `path_mode` ("Simple"|"Regex") — переключатель «Match by path» и режим Simple/Regex.
- **urls:** `domain_regex` (bool) — состояние галочки «Regex».
- Типы `ips`, `srs`, `raw` могут не использовать params.

### rule_set (для типа srs)

Массив определений rule-set'ов в формате как в `bin/wizard_template.json`: элементы с полями `tag`, `type`, `format`, `url`. При загрузке восстанавливаются в `Rule.RuleSets`; при сохранении записываются из `Rule.RuleSets`.

## Логика загрузки

1. Файл читается (state.json или выбранное именованное состояние).
2. Выполняется миграция полей при необходимости (см. миграции v1→v2 в SPECS/002-F-C-WIZARD_STATE/WIZARD_STATE_JSON_SCHEMA.md).
3. `custom_rules`: каждый элемент конвертируется в `RuleState` через **ToRuleState()**. При отсутствии или старом формате `type` тип выводится из `rule` через **DetermineRuleType(rule)**. Из `params` восстанавливается состояние UI (processes/urls); из `rule_set` — Rule.RuleSets для типа srs.
4. Модель визарда заполняется восстановленными данными.

Подробная последовательность восстановления — в коде `presenter_state.go` (LoadState, restoreCustomRules).

## Где хранится state

- **Текущее состояние:** `bin/wizard_states/state.json` (относительно ExecDir).
- **Именованные состояния:** `bin/wizard_states/<id>.json`.

Чтение/запись выполняет слой бизнес-логики (state_store); презентер создаёт состояние из модели (CreateStateFromModel) и восстанавливает модель из загруженного файла (LoadState).

## Миграции

- **v1 → v2:** `selectable_rule_states` и `custom_rules` приводятся к новому формату (см. WIZARD_STATE_JSON_SCHEMA.md). Поле `type` в custom_rules при загрузке может быть в старом виде — тогда тип выводится из `rule`.

См. также: **docs/ARCHITECTURE.md** (раздел про загрузку state), **SPECS/002-F-C-WIZARD_STATE/WIZARD_STATE_JSON_SCHEMA.md**.
