# Release Notes

## Последний релиз / Latest release

**v0.8.3** — полное описание (full details): [docs/release_notes/0-8-3.md](docs/release_notes/0-8-3.md)

**v0.8.2** — полное описание (full details): [docs/release_notes/0-8-2.md](docs/release_notes/0-8-2.md)

**v0.8.1** — полное описание (full details): [docs/release_notes/0-8-1.md](docs/release_notes/0-8-1.md)

**v0.8.0** — полное описание (full details): [docs/release_notes/0-8-0.md](docs/release_notes/0-8-0.md)

<details>
<summary><b>Новое в пререлизе / Pre-release changes</b></summary>

Черновик следующего релиза (draft): [upcoming.md](docs/release_notes/upcoming.md)

**🇷🇺 Кратко (черновик):**
- **Пользовательские SRS-правила — локальное скачивание:** Для правил типа SRS в визарде добавлена локальная загрузка SRS (кнопка ⬇/🔄/✔️). При наличии локальных файлов `bin/rule-sets/<tag>.srs` конфиг использует `type: "local"` с `path` вместо remote `url`.
- **Рефакторинг Custom Rule (типы, Raw, SRS, params):** Типы правил — константы (ips, urls, processes, srs, raw). Диалог Add/Edit Rule: название над вкладками Form/Raw; режимы Domains (Exact/Suffix/Keyword/Regex); тип SRS с подсказкой runetfreedom; при Raw→Form восстанавливаются outbound и поля по типу. Состояние UI в params.
- **Правило Processes — Match by path:** В диалоге Add/Edit Rule для типа «Processes» можно включить «Match by path» и задавать сопоставление по пути процесса (regex), а не по имени. Режим Simple: подстановка `*` как «любая последовательность» (например `*/steam/*`). Режим Regex: полные регулярные выражения. В конфиг записывается `process_path_regex` (sing-box 1.10+).
- **Кнопка перезапуска:** На дашборде Core между Start и Stop — кнопка перезапуска (🔄). Завершает процесс sing-box, вотчер поднимает снова; в UI кратко «Restarting...», смена состояния кнопок, затем «Running».
- **Сохранение в визарде:** Только запись файлов и Update (без перезапуска sing-box). Конфиг валидируется через `sing-box check` по временному `config-check.json` до записи в `config.json`; при ошибке — сообщение пользователю, рабочий конфиг не перезаписывается. Clash API перечитывается из `config.json` только при запуске sing-box.
- **Диалог Linux capabilities (issue #34):** Команда setcap в выделяемом поле, кнопка «Copy» в буфер обмена.
- **Автозапуск из Планировщика заданий:** Рекомендация при зависании: включить задержку триггера «При входе в систему» (30 с или 1 мин).
- **Визард: платформа win7:** В шаблоне визарда при сборке Win7 (GOARCH=386) применяются секции `"platforms": ["windows"]` и `"platforms": ["win7"]`.
- **Win7 CI:** Сборка Win7 (job `build-win7`) использует `go.win7.mod` с `golang.org/x/sys v0.25.0` (Go 1.20). Артефакт `singbox-launcher-<version>-win7-32.zip` стабильно попадает в release.

**🇬🇧 Summary (draft):**
- **Custom SRS rules — local download:** Custom rules of type SRS now support local SRS downloads in the Wizard Rules tab (⬇/🔄/✔️ button). When local files exist (`bin/rule-sets/<tag>.srs`), the generated config uses `type: "local"` with `path` instead of remote `url`.
- **Custom Rule refactor (types, Raw, SRS, params):** Rule type constants (ips, urls, processes, srs, raw). Add/Edit Rule: name above Form/Raw tabs; Domains mode (Exact/Suffix/Keyword/Regex); SRS type with runetfreedom hint; Raw→Form restores outbound and fields. UI state in params.
- **Processes rule — Match by path:** In Add/Edit Rule, for type «Processes» enable «Match by path» to match by process path (regex). Simple: `*` as wildcard (e.g. `*/steam/*`). Regex: full regular expressions. Stored as `process_path_regex` (sing-box 1.10+).
- **Restart button:** On Core dashboard between Start and Stop; kills sing-box, watcher restarts it; UI shows «Restarting...» then «Running».
- **Wizard save:** Write files and Update only (no sing-box restart). Config validated with `sing-box check` before overwrite; on failure user sees error. Clash API reloaded from `config.json` only when sing-box starts.
- **Linux capabilities dialog (#34):** setcap command in selectable field, «Copy» button.
- **Win7 CI:** Win7 build (job `build-win7`) uses dedicated `go.win7.mod` with pinned `golang.org/x/sys v0.25.0` (Go 1.20). Artifact `singbox-launcher-<version>-win7-32.zip` reliably included in release.

</details>
