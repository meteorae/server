# Meteorae

![GitHub Workflow Status](https://img.shields.io/github/workflow/status/meteorae/server/main) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/meteorae/server) ![GitHub release (latest by SemVer)](https://img.shields.io/github/downloads/meteorae/server/latest/total) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/meteorae/server) ![GitHub](https://img.shields.io/github/license/meteorae/server)  
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=coverage)](https://sonarcloud.io/summary/new_code?id=meteorae_server) [![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=meteorae_server) [![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=meteorae_server) [![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=ncloc)](https://sonarcloud.io/summary/new_code?id=meteorae_server)

## What is Meteorae

Meteorae is an application allowing you to serve and explore your curated media collection from anywhere.

Meteorae is built from the ground up to support every media type and allow for cross-linking metadata. For example, you might want to find more work by your favorite costume designer. Meteorae can help you figure out that they also worked on a music video by your favorite artist, or wrote a book about costume design that you might want to read.

In short, Meteorae helps you explore your content through an extensive relationship graph, making it easier to find movies, series, books or music you want to experience.

But Meteorae is also built with a specific idea of how your media server should work on a technical level. From handling very large playlists, to deeply analyzing your files or shuffling your music in the best way possible, Meteorae makes opinionated choices as to how things _should_ work.

## Building

You will need [Go](https://go.dev/doc/install) in order to build the Meteorae server.

Additional dependencies are required:

- For Debian, Ubuntu and derivatives: `libsqlite3-dev libvips-dev`
- For Fedora and derivatives: `sqlite-devel libvips-devel`

In order to build binaries, simply run `make build`.

## Contributing

If you want to help Meteorae, check out [CONTRIBUTING.md](https://github.com/meteorae/server/blob/master/CONTRIBUTING.md) to figure out how you can be a part of the project.

For the short version, here are some ways you can help:

- If you're a developer, check out [our roadmap](https://github.com/meteorae/server/blob/master/README.md) or [take on an issue](https://github.com/meteorae/server/labels/good%20first%20issue).
- If you're a translator or speak multiple languages, help us get Meteorae translated.
- You can also help [triage issues](https://github.com/meteorae/server/issues).
