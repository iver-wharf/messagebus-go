# Message bus Go library

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f31b8bb8960d49af8284c5f8c50890bf)](https://www.codacy.com/gh/iver-wharf/messagebus-go/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=iver-wharf/messagebus-go&amp;utm_campaign=Badge_Grade)
[![Go Reference](https://pkg.go.dev/badge/github.com/iver-wharf/messagebus-go)](https://pkg.go.dev/github.com/iver-wharf/messagebus-go)

Package prepared for creating and sending messages to RabbitMQ.

To use it you have to set environment variable `WHARF_INSTANCE` with proper
instance ID, such as `WHARF_INSTANCE=prod` for your production instance of
Wharf and `WHARF_INSTANCE=dev` for your development instance.

This value is used to use the same RabbitMQ instance with multiple Wharf
instances.

## Linting Golang

- Requires Node.js (npm) to be installed: <https://nodejs.org/en/download/>
- Requires Revive to be installed: <https://revive.run/>

```sh
go get -u github.com/mgechev/revive
```

```sh
npm run lint-go
```

## Linting markdown

- Requires Node.js (npm) to be installed: <https://nodejs.org/en/download/>

```sh
npm install

npm run lint-md

# Some errors can be fixed automatically. Keep in mind that this updates the
# files in place.
npm run lint-md-fix
```

## Linting

You can lint all of the above at the same time by running:

```sh
npm run lint

# Some errors can be fixed automatically. Keep in mind that this updates the
# files in place.
npm run lint-fix
```

---

Maintained by [Iver](https://www.iver.com/en).
Licensed under the [MIT license](./LICENSE).
