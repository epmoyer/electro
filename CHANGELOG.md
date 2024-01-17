# Electro Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a
Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres
to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
(None)
## 2.0.0 2023-08-03

🔴  Work in progress  

### Changed
- Default to new (dark blue) UI.
- Default to new "bar style" headings.
- Square corners on `code`.
- Bigger margins on pre and notices.
- Legacy style obtainable via mixins:
    - `doc_legacy`
        - Get pre-2.0.0 document style (old headings)
    - `ui_legacy`
        - Get pre-2.0.0 UI (lighter blue)
    - `pygments_standard`
        - Get pre-2.0.0 `pre` syntax highlighting (colors on gray)
- Convert checklist "box" items to unicode symbols:
    - `[x]` -> ✅
    - `[ ]` -> 🔲

### Added
- `mixins` system
- "Section Divider" headings.  Create them by defining a named section (in the `navigation` object of an `electro.json` config file) **without** declaring any associated documents.
### Fixed
- `menu-tree` hover color.
### TODO
- 🟡  Fenced style notices/div (`:::notice info`, and `:::no-indent`)

## 1.4.4 2023-06-30
### Changed
- Update `requirements.txt` for all missing packages. This project now works in a fresh Python 3.9 venv using `python -m pip install -r requirements.txt`.

## 1.5.0 2023-07-14
### Added
- Add theme `positron_monokai_wide_image`
## 1.4.3 2023-06-29
### Changed
- Add support for "superfences"
    - Internal:
        - `.highlight` class changed to `.search-highlight` because the `pymdownx.superfences` markdown plugin injects a `.highlight` class to identify syntax-highlighted `code` blocks (rather than the former `fenced_code` markdown plugin, which injected the `.codehilite` class).
        - Change CSS injections for pygments syntax highlighting from `.codehilite` to `.highlight` (to match the `.highlight` class injected by the `pymdownx.superfences` markdown plugin).
    - Update pygments colors for Generic.Heading (`.gh`) and Generic.Subheading (`.gu`) in `positron_monokai` style.
## 1.4.2 2023-06-17
### Fixed
- Use relative path for fonts (`fonts/{file}` instead of `\fonts/{file}`)

## 1.3.3 2023-04-09
### Fixed
- Improve (sill experimental) change bar support.
    - Bars now use anchor elements to track text position during resize and draw in their own gutter.
    - Bars can now span arbitrarily elements, or start/end within elements.
- Fix colorizing of code and pre in `positron_monokai`.

## 1.3.2 2023-04-08
### Changed
- Correctly implement code highlighting within theme(s)

### Added
- `positron_monokai` theme.

## 1.3.1 2023-04-07
### Fixed
- Replace heading links to match re-numbered headings (when auto-numbering headings)

## 1.3.0 2023-01-24
### Fixed
- Make inter-document links of the form `[](#heading-target)` work.
- For singleFile projects, highlight appropriate menu item when navigating between pages.
    - TODO: In this version, navigating to a level-3 or deeper menu item will highlight the associated document menu item but will **not** highlight the associated level-2 menu item.

## 1.2.0 2023-01-24
### Removed
- Removed the `mdx_urlize` extension due to incompatibility with markdown 3.4.1.

## 1.1.1 2022-10-10
### Changed
- Do not inject heading numbers inside fenced blocks.

## 1.1.0 2022-10-04
### Added
- Command line option: `--version` 
- Project config file option: `output_single_file`.
    - If `output_format` is `single_file`, then the (single-file) output .html document will be copied to this filename.  Must specify an .html file.  Path is relative to the project config file.
### Changed
- Ignore blank search text.
- Positron (default) theme:
    - Refine heading sizes. h2 now same as h1 (h1 has underline).
    - All headings now black (h3+ were gray).
- Adopt font `Lato` for non-heading text.
### Fixed
- Fix path to `logs/` so that app can be run without first changing directories to the app path.
- Fix `simplepack` so that it does not create working directories unless uglifying.
- Do not require `footer` in project config.
- (internal) Allow get_deprecated() to return` None`.
- Unescape embedded image file paths (e.g. `%20` for space).

## 1.0.3 2022-10-02
### Added
- `timezone` option to .json config (sets timezone for `{{% timestamp %}}` injection.)
## 1.0.2 2022-09-10
### Fixed
- Stop if a document is not found (and show fault(s) to user).
### Changed
- Search result snippets now make a best-effort to encompass the search terms.
- Single search result links in single-file mode
- Fix x-overflow pre when a #{heading} navigation is specified in URL.
- Add`overflow-x: auto;` to pre elements, and add css to show scrollbar.
## 1.0.1 2022-09-05
### Fixed
- Correct spacing bug in preformatted code blocks (padding was making it appear as though there was a blank character preceding the first character of the text).
### Changed
- Force unordered lists following paragraph text (with no intervening blank line) to be treated as the start of a new unordered list.
### Removed
- (Internal) Debug print logging

## 1.0.0 2022-09-05
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
    - `menu_level_two_number_prefix_width`: (optional) Width to use for heading number field in level 2 menu headings (e.g. `20px`)
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

