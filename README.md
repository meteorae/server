<h1 align="center">Meteorae</h1>

<p align="center">
    <img alt="GitHub Workflow Status" src="https://img.shields.io/github/workflow/status/meteorae/server/main">
    <a href="">
        <img alt="Discord" src="https://img.shields.io/discord/935381762362712084">
    </a>
    <a href="https://github.com/meteorae/server/issues">
        <img alt="GitHub issues" src="https://img.shields.io/github/issues/meteorae/server">
    </a>
    <img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/meteorae/server">
    <img alt="GitHub all releases" src="https://img.shields.io/github/downloads/meteorae/server/total">
</p>
<p align="center">
    <a href="https://sonarcloud.io/summary/new_code?id=meteorae_server">
        <img alt="Coverage" src="https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=coverage">
    </a>
    <a href="https://sonarcloud.io/summary/new_code?id=meteorae_server">
        <img alt="Maintainability Rating" src="https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=sqale_rating">
    </a>
    <a href="https://sonarcloud.io/summary/new_code?id=meteorae_server">
        <img alt="Technical Debt" src="https://sonarcloud.io/api/project_badges/measure?project=meteorae_server&metric=sqale_index">
    </a>
</p>
<p align="center">
    <img alt="GitHub" src="https://img.shields.io/github/license/meteorae/server">
    <img alt="GitHub contributors" src="https://img.shields.io/github/contributors-anon/meteorae/server">
    <a href="http://commitizen.github.io/cz-cli">
        <img alt="Commitizen friendly" src="https://img.shields.io/badge/commitizen-friendly-brightgreen.svg">
    </a>
</p>

## What is Meteorae

Meteorae is an application allowing you to serve and explore your curated media collection from anywhere.

Meteorae is built from the ground up to support every media type and allow for cross-linking metadata. For example, you might want to find more work by your favorite costume designer. Meteorae can help you figure out that they also worked on a music video by your favorite artist, or wrote a book about costume design that you might want to read.

In short, Meteorae helps you explore your content through an extensive relationship graph, making it easier to find movies, series, books or music you want to experience.

## Running

Meteorae is still in early development. While the current offering is bare-bones, we provide automated releases for Linux (Ubuntu 20.04) and Windows.

Fully-signed macOS binaries are planned, but require some exploration to figure out how to build them in our CI, while also performing the automated release process.

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
