# quanarch-go

The Go stack of [Quanarch](https://quanarch.org): the templates `qu` uses for `--lang go`.

```sh
qu new project shop --lang go --dir . --qua shop-qua --quanar shop-quanar --quan notes
```

Needs qu 0.3.6 or later and Go 1.24 or later.

## How Quanarch works in Go

- The **quanar** is one server program on `net/http`. It holds no business logic: a listener
  table listens for the quans, a store keeps their rows, and a loader brings them in.
- A **quan** is a Go package with no `main`. `Mount` asks the quanar to listen:
  `api.Get("/notes", handler)`. It cannot run on its own.
- The **qua** is the contract: a Go module with `QuanarCore`, `Init` and the thin `api` and `db`
  packages. It knows no tables, no users, no stock.

Go cannot load code into a running program, so each quan is built into its own program with the
quanar's host, which holds `main`: `go tool quan-build`, from the quan's folder. The quanar
downloads the program, checks its hash and runs it in its own process.

On Linux every quan goes through a gate before a single line of it runs, even its `init`
functions: an empty environment, its own network where the system allows it, a memory ceiling
and Landlock rules that leave it nothing but its own program. On other systems quans run
without walls, and the quanar says so when it starts.

A quan's `go tool quan-build --watch` serves its program and announces every new build on
`/events`; the quanar follows that stream, loads the quan when it opens, reloads it on a new
build and unmounts it the moment it closes.

## Templates

| Template | What it holds |
| --- | --- |
| `templates/project` | The project root: `conf.qu`, `go.work`, a README and `packages/README.md`. |
| `templates/qua` | The contract: `QuanarCore` with `ApiManager` and `DbManager`, `Init`, and the thin `api` and `db` packages. |
| `templates/quanar` | The server: listeners, store (in memory), HTTP app, loader, process, the gate, the host and `cmd/quan-build`. Also renders a **qmock**, on port 3001. |
| `templates/quan` | A quan that listens to `GET` and `POST /<name>` and stores rows in the `<name>` table. |

The parts are held together by the project's `go.work`. How a quan or a qua added later is
taken in by the quanar is declared in `stack.qu`, under `on quan` and `on qua`.

## License

Quanarch was created by [Can Berk Coşar](https://berkcosar.com) and is developed at
[Relteco](https://relteco.com).

Licensed under the [Apache License, Version 2.0](LICENSE). See [NOTICE](NOTICE).
