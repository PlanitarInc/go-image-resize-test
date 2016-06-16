The project was used to test and compare different GOlang solutions for JPEG
resizing. Among the ones we've tested are:

 - https://github.com/disintegration/imaging
 - https://github.com/nfnt/resize
 - https://github.com/bamiaux/rez

JPEG backends used in the project:
 - golang standard JPEG backend
 - golang bindings to libjpeg-turbo: https://github.com/pixiv/go-libjpeg

### Metrics

We're measuring the following metrics:
 - final image size
 - time required to resize (including decoding and encoding)
 - bytes allocated during resize (golang VM stats)
 - number of malloc operations performed during resize (golang VM stats)

### Usage

In order to disable a resizer, just comment a line with its name in
`config.yaml`. In order to enable it, uncomment the line.

```
Usage:
  resize [command]

Available Commands:
  list        List the available resizers
  run
  bench       Clean the tmp directory
  clean       Clean the tmp directory

Flags:
  -c, --config-file="./config.yaml": config file, e.g. ./config.yaml
  -H, --height=0: height of the target image
      --native[=false]: use native jpeg
  -o, --out="./tmp": output dir name
  -W, --width=0: width of the target image
```

### Examples

Compare multiple backends by resizing a particular image to 400px:

```sh
go run *.go run -W 400 <PATH-TO-IMAGE>
```

Same stuff but using golang standard JPEG decoder:

```sh
go run *.go run -W 400 <PATH-TO-IMAGE> --native
```

Run benchmark tests by resizing a particular image to 400px multiple times
and then showing the averaged metrics:

```
go run *.go bench -W 400 <PATH-TO_IMAGE>
```

Run tests on multiple images:

```
for im in images/*; do go run *.go run $im -W 400; done
```
