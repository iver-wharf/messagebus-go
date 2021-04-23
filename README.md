# Message bus Go library

[![Go Reference](https://pkg.go.dev/badge/github.com/iver-wharf/messagebus-go)](https://pkg.go.dev/github.com/iver-wharf/messagebus-go)

Package prepared for creating and sending messages to RabbitMQ.

To use it you have to set environment variable `WHARF_INSTANCE` with proper
instance ID, such as `WHARF_INSTANCE=prod` for your production instance of
Wharf and `WHARF_INSTANCE=dev` for your development instance.

This value is used to use the same RabbitMQ instance with multiple Wharf
instances.

## Linting markdown

Requires Node.js (npm) to be installed: <https://nodejs.org/en/download/>

```sh
npm install

npm run lint

# Some errors can be fixed automatically. Keep in mind that this updates the
# files in place.
npm run lint-fix
```

---

Maintained by [Iver](https://www.iver.com/en).
Licensed under the [MIT license](./LICENSE).
