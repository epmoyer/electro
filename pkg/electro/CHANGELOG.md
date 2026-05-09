# Changelog, Electro

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## v3.5.0 - 2026-05-08
### Fixed
- Strip out placeholders of the form `{{% placeholder %}}` from the HTML before adding html to the search index.
    - Without this apps like Dynamo that replace strings like `{{% doc_control_stamp_here %}}` AFTER HTML rendering would end up targeting occurrences of `{{% doc_control_stamp_here %}}` in the SEARCH INDEX, which can cause unpredictable behavior.

## v3.4.0 - 2026-04-26
### Added
- Dark gray border around images (so that you can see the edges of screenshots that are subsections of white windows)

## v3.3.0 - 2026-01-31
### Added
- Support d2 diagrams in fenced code blocks with `d2` language tag.
- Support for `@field{<fieldName>}` directives to insert dynamic content.
    - Supported fields:
        - `app_version` (e.g. "v3.3.0")
        - `app_name` (e.g. "electro")
        - `datetime_now`:  directive to insert the current date and time in the form `2026-01-31T08:58:09-08:00`.

## v3.2.0 - 2025-12-28
### Changed
- Style pre blocks to:
    - Force background color (over goldmark's inline style).
    - Auto overflow-x scroll.

## v3.1.0 - 2025-12-19
### Added
- Support new notice syntax using `@block{<noticeType>}` and `@block{end}` directives, while maintaining compatibility with legacy syntax.

## v3.0.0 - 2025-11-23
Migrated to Go (from Python). This is the initial stable release of Electro in Go.