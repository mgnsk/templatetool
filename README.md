### Usage

`go install github.com/mgnsk/templatetool`

Set `TPL_GLOB` to the glob of templates.

### Examples

Example templates defined in `example.tpl`

```
$ TPL_GLOB=*.tpl go run main.go
Template lister and renderer.

Usage:
   [command]

Available Commands:
  MapTemplate 
  Template1   
  Template2   
  help        Help about any command

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```

List variables in a template:
```
$ TPL_GLOB=*.tpl go run main.go Template1
Error: required flag(s) "MyVar1" not set
Usage:
   Template1 [flags]

Flags:
      --MyVar1 string   MyVar1
  -h, --help            help for Template1

required flag(s) "MyVar1" not set
exit status 1
```

To render, set the variables:
```
$ TPL_GLOB=*.tpl go run main.go Template1 --MyVar1 test
test
```
If a template contains array/slice/map variables, a JSON payload can be passed:
```
$ TPL_GLOB=*.tpl go run main.go MapTemplate
Error: required flag(s) "Header", "Lists" not set
Usage:
   MapTemplate [flags]

Flags:
      --Header string   String
      --Lists string    JSON
  -h, --help            help for MapTemplate

required flag(s) "Header", "Lists" not set
exit status 1
```

Rendering:
```
$ TPL_GLOB=*.tpl go run main.go MapTemplate --Header "<h2>Sections</h2>" --Lists "$(cat <<JSON
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
<h2>Sections</h2>

<section>
    First section:
    <p>First item</p>
    <p>Second item</p>
</section>

<section>
    Second section:
    <p>item1</p>
    <p>item2</p>
</section>
```


TODO: streaming JSON from stdin for rendering a single template with multiple data.