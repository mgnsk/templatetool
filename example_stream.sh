#!/bin/bash

trap onerr ERR
onerr(){ while caller $((n++)); do :; done; }

# Streaming mode: input is a JSON array of objects that define the whole template.

STREAM=1 TPL_GLOB=*.tpl go run main.go MapTemplate <<JSON
[
    {
        "Header": "Header1",
        "Lists": {
            "First section": [
                "First item",
                "Second item"
            ],
            "Second section": [
                "item1",
                "item2"
            ]
        }
    },

    {
        "Header": "Header2",
        "Lists": {
            "First section2": [
                "First item2",
                "Second item2"
            ],
            "Second section2": [
                "item12",
                "item22"
            ]
        }
    },

    {
        "Header": "Header3",
        "Lists": {
            "First section3": [
                "First item3",
                "Second item3"
            ],
            "Second section3": [
                "item13",
                "item23"
            ]
        }
    }
]
JSON
