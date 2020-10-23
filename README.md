### Usage

`go get github.com/mgnsk/templatetool`

Set `TPL_GLOB` to the glob of templates.
`STREAM=1` can be used to render a template with data from a json stream.

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

See `example_once.sh` and `example_stream.sh` to see how to deal with  array/slice/map variables or JSON streaming from standard input.
