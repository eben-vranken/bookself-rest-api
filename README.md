# Bookshelf REST API (in-memory)

A JSON REST API for managing a collection of books. Full CRUD on `/` with state held in a single in-memory `map[int]Book` behind a `sync.RWMutex`. No database, no router library — just `net/http` and Go 1.22+ pattern matching.

## How to use

```sh
go run .
```

Server listens on `127.0.0.1:8080`.

| Method   | Path     | What it does                                  |
| -------- | -------- | --------------------------------------------- |
| `GET`    | `/`      | List every book as a JSON object keyed by id  |
| `GET`    | `/{id}`  | Fetch one book by id                          |
| `POST`   | `/`      | Create a book; server assigns the id          |
| `PUT`    | `/{id}`  | Replace the book at that id                   |
| `DELETE` | `/{id}`  | Remove the book at that id                    |

```sh
$ curl -X POST http://localhost:8080/ \
    -H 'Content-Type: application/json' \
    -d '{"title":"The Pragmatic Programmer","author":"Hunt & Thomas","release_date":"1999-10-20T00:00:00Z"}'

$ curl http://localhost:8080/0

$ curl -X DELETE http://localhost:8080/0
```

## What I learned

- **An HTTP server is concurrent by default.** `net/http` spawns a goroutine per request. Two requests touching the same `map` is a data race; the mutex isn't optional.
- **`RWMutex` only helps when reads dominate.** Multiple readers can hold `RLock` at once; writers serialize. For a read-heavy CRUD API, that's the right shape.
- **JSON struct tags are picky about whitespace.** `` `json:"id"` `` works; `` `json: "id"` `` silently fails and serializes under the Go field name. Parsed by reflection at runtime — the compiler doesn't catch it.
- **`len(map) + 1` is not an ID generator.** Breaks the moment anything is deleted. Either scan for the max and add one, or store the next-id separately.
- **The comma-ok map idiom is the only correct existence check.** `_, ok := books[id]`. Comparing against `Book{}` couples the check to the struct shape and misclassifies genuinely empty records as missing.
- **PUT requires the URL to win over the body.** Otherwise a client can repoint records by sending a different `id` in the JSON. Overwrite `book.Id = idToUpdate` before storing.
- **Set status *before* writing the body.** `w.Write(...)` implicitly calls `WriteHeader(200)`. Once flushed, headers can't be changed and a later `WriteHeader(404)` is silently ignored.
- **`http.Error` is a one-liner replacement for the three-step error response.** Writes `Content-Type: text/plain`, the status code, and the message in one call.
- **`delete(m, k)` can't fail on a missing key** — it's a no-op. The existence check has to happen *before* the delete if you want a 404 for missing IDs.
- **`req.PathValue("id")` returns a string.** Run it through `strconv.Atoi` and 400 on a non-numeric input before touching the map.
