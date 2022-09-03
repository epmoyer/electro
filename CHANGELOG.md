# Electro Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a
Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres
to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
(None)
## 1.0.0 2022-09-01
### Changed
- Accept a project file (.json) or dir as command line argument
- Restyle positron theme for "general case" use.
### Added
- Project .json keys:
    - `output_format`: One of "single_file" or "static_site"
    - `enable_newline_to_break`: Optional. If `true`, then newlines in .md will be rendered as `<br>`
    - `strip_frontmatter`: If `true`, strip frontmatter from .md documents (i.e. `---` enclosed block)
    - `number_headings`: If `true`, add numbers to headings.
    - `number_headings_at_level`: If `number_headings` is true, then this controls the heading level at which numbering will begin. This defaults to `1`, but typically a user will set this to `2`if they have only one level 1 heading, and that heading is the title of the document.
## Removed
- Command line `--nobreak` option

## 0.1.7 2022-08-06
### Fixed
-  Refine heading_text_to_id (replace unicode NBSP with spaces)

## 0.1.6 2022-08-06
### Fixed
-  Refine heading_text_to_id 

## 0.1.5 2022-08-06
### Fixed
-  Refine heading_text_to_id 

## 0.1.4 2022-08-04
### Fixed
- Scroll to top of page when clicking on top level item in sidebar.

## 0.1.3 2022-05-56
### Added
- Add `--nobreak` option (to disable nl2br)
- Use `overlay.css` if provided in a project's `docs/` directory.

## 0.1.2 2022-02-16
### Changed
- New favicon

## 0.1.1 2022-02-16
### Changed
- Set page `title` tag.
- Use a placeholder `favicon.ico`.
- Remove `\` from menu `href`.
- Remove `\` from search `href`.

## 0.1.0 2022-02-16
(Initial)

