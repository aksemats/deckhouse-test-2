#!/bin/bash

source /antiopa/shell_lib.sh

function __config__() {
  echo '
{
  "beforeHelm": 5
}'
}

function __main__() {
  if ! kubectl -n kube-system get secret ingress-conversion-webhook > /dev/null 2> /dev/null ; then
    ca=$(jo CN=ingress-conversion-webhook key="$(jo algo=ecdsa size=256)" ca="$(jo expiry=87600h)" | cfssl gencert -initca -)
    ca_cert=$(echo "$ca" | jq .cert -r)

    # Создадим конфиг для cfssl gencert
    config='{"signing":{"default":{"expiry":"87600h","usages":["signing","key encipherment","requestheader-client"]}}}'

    cert=$(jo CN=ingress-conversion-webhook hosts="$(jo -a ingress-conversion-webhook.kube-system ingress-conversion-webhook.kube-system.svc)" key="$(jo algo=ecdsa size=256)" | cfssl gencert -ca=<(echo $ca | jq .cert -r) -ca-key=<(echo $ca | jq .key -r) -config=<(echo $config) -)
    cert_pem=$(echo "$cert" | jq .cert -r)
    cert_key=$(echo "$cert" | jq .key -r)
  else
    cert=$(kubectl -n kube-system get secret ingress-conversion-webhook -o json)
    ca_cert=$(echo "$cert" | jq -r '.data."webhook-ca.crt"' | base64 -d)
    cert_pem=$(echo "$cert" | jq -r '.data."tls.crt"' | base64 -d)
    cert_key=$(echo "$cert" | jq -r '.data."tls.key"' | base64 -d)
  fi

  values::set nginxIngress.ingressConversionWebhookCa "$ca_cert"
  values::set nginxIngress.ingressConversionWebhookPem "$cert_pem"
  values::set nginxIngress.ingressConversionWebhookKey "$cert_key"
}

hook::run "$@"
