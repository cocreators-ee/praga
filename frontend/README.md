# praga-frontend

Built on SvelteKit. Depends on `pnpm`, tested on version 9.0.6 and Node 20.12.2.

Check root README.md for most of the details but for development run:

```shell
pnpm run dev
```

To build, run:

```shell
pnpm install
pnpm run build
```

The proxy configuration in `vite.config.ts` should route `/api` requests to the backend that should be
listening to http://localhost:8086
