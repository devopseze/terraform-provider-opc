#!/usr/bin/env bash

docs=$(ls website/docs/**/*.markdown)
error=false

for doc in $docs; do
  dirname=$(dirname "$doc")
  category=$(basename "$dirname")


  case "$category" in
    "guides")
      # Guides require a page_title
      if ! grep "^page_title: " "$doc" > /dev/null; then
        echo "Guide is missing a page_title: $doc"
        error=true
      fi
      ;;

    "d")
      # no data sources subcategories
    ;;

    "r")
      # Resources and data sources require a subcategory
      if ! grep "^subcategory: " "$doc" > /dev/null; then
        echo "Doc is missing a subcategory: $doc"
        error=true
      fi
      ;;

    *)
      error=true
      echo "Unknown category \"$category\". " \
        "Docs can only exist in r/, d/, or guides/ folders."
      ;;
  esac
done

if $error; then
  exit 1
fi

exit 0
