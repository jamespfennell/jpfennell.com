# Node status pages

Each of my servers serves a status page at its own hostname
    (e.g. [a.jamesfennell.dev](https://a.jamesfennell.dev))
    listing what runs on it.

The page has the same design as the rest of the website:
    Hugo renders `layouts/node-template/single.html` using the site's base layout
    into a Go template (with `[[ ... ]]` delimiters, so they don't clash with Hugo's).
The `node` binary in this directory fills in that template from a config file
    on each request, so config changes don't need a restart.
The Docker image (`--target node`) contains the binary, the template and the stylesheets.

The config files live in the
    [server-config](https://github.com/jamespfennell/server-config) repo.
The full spec is the `Config` type in `main.go`.
Entries can live in their own files, listed under `include` in the main config;
    each included file contains a list of entries.

To run locally:

    hugo -d /tmp/public
    go run . \
        --config=$PATH_TO_CONFIG_FILE \
        --template=/tmp/public/node-template/index.html \
        --static=/tmp/public \
        --addr=localhost:8000
