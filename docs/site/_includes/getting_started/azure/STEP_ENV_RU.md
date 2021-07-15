Чтобы **Deckhouse Platform {% if page.revision == 'ee' %}Enterprise Edition{% else %}Community Edition{% endif %}** смог управлять ресурсами в облаке Microsoft Azure, необходимо создать сервисный аккаунт. Подробная инструкция по этому действию доступна в [документации провайдера](https://cloud.google.com/iam/docs/service-accounts). Здесь мы представим краткую последовательность действий, которую необходимо выполнить с помощью консольной утилиты Azure CLI:
- Установите [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) и выполните `login`;
- Экспортируйте переменную окружения, подставив вместо значения `my-subscription-id` идентификатор подписки Amazon AWS:
  ```shell
  export SUBSCRIPTION_ID="my-subscription-id"
  ```
- Создайте service account, выполнив команду:
  ```shell
  az ad sp create-for-rbac --role="Contributor" --scopes="/subscriptions/$SUBSCRIPTION_ID" --name "account_name"
  ```