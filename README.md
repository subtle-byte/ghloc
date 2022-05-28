# GitHub Lines Of Code

It is just for fun project for counting the number of non-empty lines of code in any public Github repository.

The idea is simple: you make a request to the API in the format `/<username>/<repository>/<branch>` (or just `/<username>/<repository>` (the branch `master` or `main` will be used if any exists)) and you get the response with human-readable JSON.

It is deployed on the [ghloc.bytes.pw](http://ghloc.bytes.pw) (although no any guaranty), so it posible to get statistics using [ghloc.bytes.pw/go-chi/chi](http://ghloc.bytes.pw/go-chi/chi) for example.

You can show only some files using `match` URL parameter, e.g. with `/someuser/somerepo?match=js` only paths containing `js` will be considered. Examples of more powerful usage:
* `match=.js$` will show only paths ending with `.js`.
* `match=^src/` will show only paths starting with `src/` (i.e. placed in the `src` folder).
* `match=!test` will filter out paths containing `test`.
* `match=!test,!.sum` will filter paths containing `test` or `.sum`.
* `match=.json$,!^package-lock.json$` will show only json files except for `package-lock.json` file.

There is also `filter` URL parameter, which has the opposite behavior to `match` parameter. `filter` has the same syntax but it declares which files must be filtered out.

### TODO

* Use `context.Context`.
* Add the prioritized tasks-queue for uncached requests? Limited number of the tasks are executed concurrently.
* Investigate impact on performance of the fact the repositories returns large slices.
