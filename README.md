# GitHub Lines Of Code

It is just for fun project for counting the number of lines of code in any public Github repository.

The idea is simple: you make a request to API endpoint in the format `/<username>/<repository>/<branch>` (or just `/<username>/<repository>` (the branch `master` or `main` will be used if any exists)) and you get the response with human-readable JSON.

It is deployed on the [ghloc.bytes.pw](http://ghloc.bytes.pw) (although no any guaranty), so it posible to get statistics using [ghloc.bytes.pw/go-chi/chi](http://ghloc.bytes.pw/go-chi/chi) for example.

You can filter some files from the results using `filter` URL parameter, e.g. `/someuser/somerepo?filter=test` will ignore all paths containing `test`. Examples of more powerful usage:
* `filter=test,.sum` will filter paths containing `test` or `.sum`.
* `filter=_test.go$,^docs/` will filter paths ending with `_test.go` or starting with `docs/`.
* `filter=.md$,!^README.md$` will filter all markdown files (i.e. ending with `.md`) except file `README.md` in the root of the repository.
* `filter=` will filter all paths (you will get zero result, it is because the filter is empty string and any path contains empty string).
* `filter=,!.go` will filter all paths, except ones containing `.go`.
* `filter=,!.go$` will filter all paths, except ones ending with `.go`.
* `filter=,!^src/` will filter all paths, except ones starting with `src/` (i.e. placed in the `src` folder).
