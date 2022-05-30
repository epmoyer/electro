#!/usr/bin/env python

from pathlib import Path
import re


def make_html_icons_inline(path_file_in, path_file_out):
    """Take an HTML file and write a new version with inline Base64 encoded icons."""
    path_base = path_file_in.parent
    out_lines = []
    with open(path_file_in, "r") as file:
        in_lines = [line.strip('\n') for line in file.readlines()]
    for line in in_lines:
        if re.match(r'.*rel=["\']shortcut icon["\']', line):
            out_lines.append(convert_icon(path_base, line))
        else:
            out_lines.append(line)

    with open(path_file_out, 'w') as file:
        file.write('\n'.join(out_lines))


def convert_icon(path_base, line):
    print(f' 🔵  Converting icon in line : {line}')
    href = re.findall(r'href=["\'](.*?)["\']', line)
    href_expression = re.findall(r'href=["\'].*?["\']', line)
    print(f' {href=} {href_expression=}')
    if len(href) != 1 or len(href_expression) != 1:
        print(
            f'🔴 Expected to find 1 and only 1 href entry on line: "{line}". '
            'Skipping icon inline conversion.'
        )
        return line
    # Extract the (single) regex search results
    href = href[0]
    href_expression = href_expression[0]

    if href.startswith('/'):
        # Strip the leading url, so we can use it as a local path
        href = href[1:]

    path_icon = path_base / Path(href)
    if not path_icon.exists() and path_icon.is_file():
        print(f'🔴 No icon file found at href {href}. Skipping icon inline conversion.')
        return line
    icon_base64 = file_to_base64(path_icon)
    href_inline = f'href="data:image/x-icon;charset=utf-8;base64,{icon_base64}"'

    return line.replace(href_expression, href_inline)


def file_to_base64(path_file):
    """
    Returns the content of a file as a Base64 encoded string.
    :param path_file: Path to the file.
    :type path_file: str
    :return: The file content, Base64 encoded.
    :rtype: str
    """
    import base64

    with open(path_file, 'rb') as f:
        encoded_str = base64.b64encode(f.read())
    return encoded_str.decode('utf-8')

