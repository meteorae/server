# Meteorae

![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/meteorae/server) ![GitHub branch checks state](https://img.shields.io/github/checks-status/meteorae/server/master) ![GitHub release (latest by SemVer)](https://img.shields.io/github/downloads/meteorae/server/latest/total) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/meteorae/server) ![GitHub](https://img.shields.io/github/license/meteorae/server)

## What is Meteorae

Meteorae is an application allowing you to serve and explore your curated media collection from anywhere.

Meteorae is built from the ground up to support every media type and allow for cross-linking metadata. For example, you might want to find more work by your favorite costume designer. Meteorae can help you figure out that they also worked on a music video by your favorite artist, or wrote a book about costume design that you might want to read.

In short, Meteorae helps you explore your content through an extensive relationship graph, making it easier to find movies, series, books or music you want to experience.

But Meteorae is also built with a specific idea of how your media server should work on a technical level. From handling very large playlists, to deeply analyzing your files or shuffling your music in the best way possible, Meteorae makes opinionated choices as to how things _should_ work.

## Building

You will need [Go](https://go.dev/doc/install) in order to build the Meteorae server.

Additional dependencies are required:

- For Debian, Ubuntu and derivatives: `libsqlite3-dev libmagickcore-dev libmagickwand-dev`
- For Fedora and derivatives: `sqlite-devel ImageMagick-devel`

In order to build binaries, simply run `make build`.

## Contributing

If you want to help Meteorae, check out [CONTRIBUTING.md](https://github.com/meteorae/server/blob/master/CONTRIBUTING.md) to figure out how you can be a part of the project.

For the short version, here are some ways you can help:

- If you're a developer, check out [our roadmap](https://github.com/meteorae/server/blob/master/README.md) or [take on an issue](https://github.com/meteorae/server/labels/good%20first%20issue).
- If you're a translator or speak multiple languages, help us get Meteorae translated.
- You can also help [triage issues](https://github.com/meteorae/server/issues).
