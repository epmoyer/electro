# Changelog, Electro

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## v3.10.0 - 2026-05-21
### Added
- Number appendix sections as "A.1", "A.2", etc. if the section title starts with "Appendix" (case-insensitive).

## v3.9.0 - 2026-05-21
### Added
- CSS print directives to control page margins and to not break table rows across page boundaries.
    - Safri ignores the tr break directive, but Edge/Chrome print to PDF very nicely, even replicating table heaings when crossing to a new page.

## v3.8.0 - 2026-05-21
### Added
- Preserve existing styles for tables when printing.
- Pragma for table cell background color by content partial match.
  - Example: `@pragma{table_cell_bg_color_by_content_partial:foo, #ffff00}`

### Changed
- All table cell background color pragmas now do a case-insensitive match.
  - `table_cell_bg_color_by_content`
  - `table_cell_bg_color_by_content_partial`

## v3.7.0 - 2026-05-21
### Added
- Add table cell background color control via pragmas.
  - Examples: 
    - `@pragma{table_cell_bg_color_by_content:pass, #d0ffd0}`
    - `@pragma{table_cell_bg_color_by_content:fail, #ffd0d0}`
    - `@pragma{table_cell_bg_color_clear_all}`

## v3.6.0 - 2026-05-21
### Added
- Add `@table{attachments/filename.csv}` directive to embed a formatted table from a CSV file.

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
