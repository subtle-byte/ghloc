# ghloc (GitHub Lines Of Code)

It is just for fun project for counting the number of non-empty lines of code in a project.

It can work in 2 modes:
* As console utility for getting info about current directory.
* As server for getting info about any public Github repository.

## CLI mode

Installation (it uses `go` tool):
```shell
go install github.com/subtle-byte/ghloc/cmd/ghloc@latest
```

And then to count lines of code in the current directory - run the command `ghloc`. The web page will be open with the results, e.g.:

<img src="https://user-images.githubusercontent.com/71576382/230733010-c740aa8b-fb66-4016-ac5c-1d946c5e733a.png" width="600">

You also can print results in the console using `ghloc -c`. Also if you want to count only some files you can use `-m` (stands for matcher), e.g. to consider only Markdown files use `ghloc -m .md` (see full matcher syntax below).

Thanks @pajecawav for this web UI (https://github.com/pajecawav/ghloc-cli-ui).

## Server mode

The idea is simple: you make a request to the API in the format `/<username>/<repository>/<branch>` (or just `/<username>/<repository>` (the branch `master` or `main` will be used if any exists)) and you get the response with human-readable JSON.

It is deployed on the [ghloc.ifels.dev](https://ghloc.ifels.dev) (although no any guaranty), so it possible to get statistics using [ghloc.ifels.dev/go-chi/chi](http://ghloc.ifels.dev/go-chi/chi) for example.

You can see only some files using `match` URL parameter, e.g. with `/someuser/somerepo?match=js` only paths containing `js` will be considered. Examples of more powerful usage:
* `match=.js$` will show only paths ending with `.js`.
* `match=^src/` will show only paths starting with `src/` (i.e. placed in the `src` folder).
* `match=!test` will filter out paths containing `test`.
* `match=!test,!.sum` will filter out paths containing `test` or `.sum`.
* `match=.json$,!^package-lock.json$` will show only json files except for `package-lock.json` file.

There is also `filter` URL parameter, which has the opposite behavior to `match` parameter. `filter` has the same syntax but it declares which files must be filtered out.

To make the response more compact you can use `pretty=false`, e.g. `/someuser/somerepo?pretty=false`.

There is useful web frontend for this API: https://github.com/pajecawav/ghloc-web (thanks @pajecawav).

## TODO

* Use `context.Context`.
* Add the prioritized tasks-queue for uncached requests? Limited number of the tasks are executed concurrently.
* Investigate impact on performance of the fact the repositories returns large slices.
