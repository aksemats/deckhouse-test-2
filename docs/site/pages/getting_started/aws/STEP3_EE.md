---
title: "Deckhouse Platform EE in Amazon AWS: Preparing environment"
title_main: "Deckhouse Platform EE in Amazon AWS"
title_sub: "Preparing environment"
permalink: en/gs/aws/step3-ee.html
layout: page-nosidebar-notitle
revision: ee
platform_code: aws
platform_name: Amazon AWS
lang: en
toc: false
---

<link rel="stylesheet" type="text/css" href='{{ assets["getting-started.css"].digest_path }}' />
{::options parse_block_html="false" /}

<h1 class="docs__title">{{ page.title_main }}</h1>
{% include getting_started/steps.html steps=9 step=3 title=page.title_sub %}

{% include getting_started/request_ee_access.md %}

{% include getting_started/aws/STEP_ENV.md %}

{% include getting_started/buttons.html step=3 nextStepCaption='preparingConfig' %}