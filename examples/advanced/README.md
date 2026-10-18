# shop · Go advanced example

A full-stack Quanarch application: a quanar server, a qua, three quans and a React client.

```
shop/
├── go.work              every Go module of the shop, together
├── packages/shop-qua/   the qua: definitions, types and thin calls; no logic
├── quanar/              the server: managers in internal/, each quan in its own process
├── quans/
│   ├── users/           accounts, roles, sign-in; asks the quanar for tokens
│   ├── products/        catalogue and stock; reserves stock for orders
│   └── orders/          orders; follows products by events
├── artifacts/           built quans, served and announced on /events/<quan>
├── tools/               artifacts, build (and seal), register, start, check
└── web/                 the React client; it only knows the quanar, under /api
```

## Run

Requires Go 1.24 or later and Node.js 24 or later (for the web client).

```sh
npm install --prefix web     # once: the web client
go run ./tools/cmd/start     # builds and seals the quans, starts everything
```

| Part | Address |
| --- | --- |
| web | http://localhost:5180 |
| quanar | http://localhost:3000 (`/health`, `/quans`) |
| artifacts | http://localhost:4100 |

The first start creates `admin@shop.test` / `admin1234`. Ctrl+C stops all three.

If a port is taken, move the quanar and tell the web where it went:

```sh
PORT=3200 QUANAR_URL=http://localhost:3200 go run ./tools/cmd/start    # ARTIFACTS_PORT moves the artifacts
```

Each part on its own, in three terminals (after `go run ./tools/cmd/build`):

```sh
go run ./tools/cmd/artifacts     # 1 · artifacts
cd quanar && go run .            # 2 · quanar
cd web && npm run dev            # 3 · web
```

Working on a quan: `go run ./tools/cmd/register --open` lists the quans without hashes, then
`cd quans/orders && go tool quan-build --watch --serve=false --out ../../artifacts/quans/orders/orders`
rebuilds on every change; the artifacts store announces it and the quanar takes the new build
in at once. `start` seals them again.

| Command | |
| --- | --- |
| `go run ./tools/cmd/build` | build every quan into the artifacts and seal them |
| `go run ./tools/cmd/start` | `build`, then start artifacts, quanar and web |
| `go run ./tools/cmd/register` | seal the built quans into `quanar/data/quans.json`; `--open` without hashes |
| `go run ./tools/cmd/check` | `go vet` every module, type-check the web |

The database is `quanar/data/shop.db`; delete it to start over.

## A quan declares, the quanar does

A quan has no `main`. It writes no route, no table and no check by hand: its package says what it
is in one `var Manifest = shopqua.QuanManifest{...}`. `go tool quan-build` links it with the
quanar's host, and when the quanar loads it, its managers make the declarations real:

| The quan declares | in | The quanar's manager |
| --- | --- | --- |
| models: fields, types, unique, index, enum | `models.go` | `internal/schema` creates and migrates the tables |
| permissions | `permissions.go` | `internal/permissions` knows them; tokens say who holds them |
| apis: endpoint, method, middlewares, permissions, validation, handler | `apis.go` | `internal/apis` opens them, per quan |
| validation rules | `validations.go` | `internal/validation` answers 422 before the quan is reached |
| events it listens to | `events.go` | `internal/events` carries them between quans |

`internal/registry` runs them in that order, then lets the quan start. A request passes the
quanar's middlewares (`authenticate`, `rate-limit`), its permission check and validation, then
the quan's own middlewares (`middlewares.go`), then the handler. Database calls, password hashing,
tokens and events go back to the quanar through the qua. Rows come back as the quan's own types:
`db.Model[Product]("products")`.

## An order, by events

`orders` never reads the products' tables. `POST /orders` stores a pending order and emits
`orders.placed`. `products` takes the stock in one transaction and answers `products.reserved`
or `products.rejected`; `orders` follows. Cancelling emits `orders.cancelled` and `products`
gives the stock back.

## Walls

On Linux each quan goes through a gate before a single line of it runs: an empty environment,
its own network, a memory ceiling and Landlock rules that leave it nothing but its own program.
On other systems the quans run without walls, and the quanar says so when it starts.

## Try

- Sign in as a customer and add a product: 403, the quanar checked `products:manage`.
- Send a product with a negative price: 422 with every issue, before the quan runs.
- Order more than the stock: the order is rejected, nothing is taken.
- Change a program in `artifacts/quans` by hand: its hash no longer matches and the quan is unmounted.
- `curl localhost:3000/quans` lists every quan's apis, models, permissions and events.
