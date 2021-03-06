#!/bin/bash

#
# Update configuration.html page for modules from the corresponding module openapi schema
#

for schema_path in $(find $MODULES_DIR -regex '^.*/openapi/config-values.yaml$' -print); do
  module_path=$(echo $schema_path | cut -d\/ -f-2 )
  module_name=$(echo $schema_path | cut -d\/ -f2 | cut -d- -f2-)
  mkdir -p _data/schemas/${module_name}
  cp -f $schema_path _data/schemas/${module_name}/
  if [ -f $module_path/openapi/doc-ru-config-values.yaml ]; then
     echo -e "\ni18n:\n  ru:" >>_data/schemas/${module_name}/config-values.yaml
     cat $module_path/openapi/doc-ru-config-values.yaml | sed 's/^/    /' >>_data/schemas/${module_name}/config-values.yaml
  fi
  if [ ! -f ${module_path}/docs/CONFIGURATION.md ]; then
      continue
  fi
  grep -q '<!-- SCHEMA -->' ${module_path}/docs/CONFIGURATION.md
  if [ $? -eq 0 ]; then
    # Apply schema
    echo "Generating schema ${schema_path} for ${module_path}/docs/CONFIGURATION.md"
    sed -i "/<!-- SCHEMA -->/i\{\{ site.data.schemas.${module_name}.config-values \| format_configuration \}\}" ${module_path}/docs/CONFIGURATION.md
  else
    echo "WARNING: Schema ${schema_path} found but there is no '<!-- SCHEMA -->' placeholder in the ${module_path}/docs/CONFIGURATION.md"
  fi
done

if [ -d /src/global ]; then
  mkdir -p /srv/jekyll-data/documentation/_data/schemas/global
  cp -f /src/global/config-values.yaml _data/schemas/global/
  echo -e "\ni18n:\n  ru:" >>_data/schemas/global/config-values.yaml
  cat /src/global/doc-ru-config-values.yaml | sed 's/^/    /' >>_data/schemas/global/config-values.yaml
fi

