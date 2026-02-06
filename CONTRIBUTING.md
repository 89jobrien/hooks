# Contributing

Thanks for contributing!

## Development

```bash
make all
make test
make config
```

## Style

- Run `gofmt` before committing.
- Use Conventional Commits (e.g., `feat: add audit hook`).

## Config changes

- Edit `config.yaml` only.
- Regenerate outputs with `make -C hooks config` from repo root.
