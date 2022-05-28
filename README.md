# ghloc (GitHub Lines Of Code)

It is just for fun project for counting the number of non-empty lines of code in a project.

It can work in 2 modes:
* As server for getting info about any public Github repository.
* As console utility for getting info about current directory.

## Server mode

The idea is simple: you make a request to the API in the format `/<username>/<repository>/<branch>` (or just `/<username>/<repository>` (the branch `master` or `main` will be used if any exists)) and you get the response with human-readable JSON.

It is deployed on the [ghloc.bytes.pw](http://ghloc.bytes.pw) (although no any guaranty), so it posible to get statistics using [ghloc.bytes.pw/go-chi/chi](http://ghloc.bytes.pw/go-chi/chi) for example.

You can show only some files using `match` URL parameter, e.g. with `/someuser/somerepo?match=js` only paths containing `js` will be considered. Examples of more powerful usage:
* `match=.js$` will show only paths ending with `.js`.
* `match=^src/` will show only paths starting with `src/` (i.e. placed in the `src` folder).
* `match=!test` will filter out paths containing `test`.
* `match=!test,!.sum` will filter paths containing `test` or `.sum`.
* `match=.json$,!^package-lock.json$` will show only json files except for `package-lock.json` file.

There is also `filter` URL parameter, which has the opposite behavior to `match` parameter. `filter` has the same syntax but it declares which files must be filtered out.

There is useful web frontend for this API: https://github.com/pajecawav/ghloc-web (thanks @pajecawav).

## CLI mode

Installation (it uses `go` tool):
```console
go install github.com/subtle-byte/ghloc/cmd/ghloc
```

And then to count lines of code in the current directory - run the command `ghloc`. The web page will be open with the results, e.g.:

<img src="https://user-images.githubusercontent.com/71576382/170814341-a5467b61-b974-4d7a-af80-043037a46608.png" width="600">

Thanks @pajecawav for this web UI (https://github.com/pajecawav/ghloc-cli-ui).

## TODO

* Use `context.Context`.
* Add the prioritized tasks-queue for uncached requests? Limited number of the tasks are executed concurrently.
* Investigate impact on performance of the fact the repositories returns large slices.
