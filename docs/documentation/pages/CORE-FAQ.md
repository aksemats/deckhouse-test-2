---
title: FAQ
permalink: en/features/core-faq.html
---

## Как узнать все параметры Deckhouse?

Все ключевые настройки Deckhouse, включая параметры модулей, хранятся в ConfigMap `deckhouse` namespace `d8-system`. Посмотреть содержимое:
```
kubectl -n d8-system get cm deckhouse -o yaml
```

## Как найти документацию по установленной у меня версии?

Документация запущенной в кластере версии Deckhouse доступна по адресу `deckhouse.<cluster_domain>`, где `<cluster_domain>` - DNS имя в соответствии с шаблоном из параметра `global.modules.publicDomainTemplate` конфигурации.

## Как установить желаемый канал обновлений?
Чтобы перейти на другой канал обновлений автоматически, минимизировав переключение версий в кластере, нужно у модуля изменить (установить) параметр `releaseChannel`. В этом случае включится механизм [автоматической стабилизации релизного канала](#как-работает-автоматическая-стабилизация-релизного-канала).

Пример конфигурации модуля:
```yaml
deckhouse: |
  releaseChannel: RockSolid
```

## Как работает автоматическая стабилизация релизного канала?
При указании в конфигурации параметра `releaseChannel`, Deckhouse сам переключит свой image на соответствующий тег Docker-образа. Дополнительных действий со стороны пользователя не требуется.

**Внимание:** переключение не происходит мгновенно и зависит от обновлений Deckhouse.

Каждые 10 минут запускается скрипт стабилизации канала обновлений, который реализует следующую логику:
* Если указанный канал обновлений соответствует тегу Docker-образа Deckhouse — ничего не произойдет;
* При смене канала обновлений на более стабильный (например с `Alpha` на `EarlyAccess`) будет произведен плавный переход:

  - Сначала проверяется равенство [digest](https://success.mirantis.com/article/images-tagging-vs-digests) для тегов Docker-образов, соответствующих текущему каналу обновлений и ближайшему к нему более стабильному (в примере — это каналы `Alpha` и `Beta`).

  - Если digest'ы равны, будет проверен следующий по очереди тег (в примере, это тэг соответствующий каналу обновлений `EarlyAccess`).

  - В результате, Deckhouse будет переключен на более стабильный канал обновлений c digest'ом, равным текущему.

* Если указан менее стабильный канал обновлений, чем тот, который соответствует текущему тегу Docker-образа Deckhouse — будет выполнена сверка digest'ов, соответствующих образам Docker для текущего канала обновлений и следующего, менее стабильного. Например, если необходимо перейти на канал `Alpha` с текущего канала `EarlyAccess`, — сравнивается `EarlyAccess` и `Beta`:

  - Если digest не равны, Deckhouse будет переключен на следующий канал обновлений (в нашем случае на `Beta`). Это необходимо, чтобы не пропустить важные миграции, которые  выполняются при обновлении Deckhouse.

  - Если digest равны, будет проверен следующий по убыванию стабильности канал обновлений (в нашем случае `Alpha`).

  - Когда проверка дойдет до желаемого канала обновлений (в примере — `Alpha`), переключение Deckhouse произойдет независимо от равенства digest.

В итоге, постоянный запуск скрипта стабилизации рано или поздно приведет Deckhouse к состоянию, при котором тег его Docker-образа будет соответствовать заданному каналу обновлений.