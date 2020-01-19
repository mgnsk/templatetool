#!/bin/bash

trap onerr ERR
onerr(){ while caller $((n++)); do :; done; }

TPL_GLOB=*.tpl go run main.go MapTemplate --Header "<h2>Sections</h2>" --Lists "$(cat <<JSON
{
    "First section": [
        "First item",
        "Second item"
    ],
    "Second section": [
        "item1",
        "item2"
    ]
}
JSON
)"
