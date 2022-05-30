#!/usr/bin/env python

from pathlib import Path
import re


def make_html_fonts_inline(path_file_in, path_file_out):
    """
    Takes an HTML file and writes a new version with inline Base64 encoded
    fonts.
    """
    path_base = path_file_in.parent
    out_lines = []
    with open(path_file_in, "r") as file:
        in_lines = [line.strip('\n') for line in file.readlines()]
    for line in in_lines:
        if 'format("woff")' in line:
            out_lines.append(convert_font(path_base, line, "woff"))
        elif 'format("woff2")' in line:
            out_lines.append(convert_font(path_base, line, "woff2"))
        else:
            out_lines.append(line)

    with open(path_file_out, 'w') as file:
        file.write('\n'.join(out_lines))


def convert_font(path_base, line, format):
    url = re.findall(r'url\("(.*?)"\)', line)
    url_expression = re.findall(r'url\(".*?"\)', line)
    if len(url) != 1 or len(url_expression) != 1:
        print(
            f'🔴 Expected to find 1 and only 1 font url entry on line: "{line}". '
            'Skipping font inline conversion.'
        )
        return line
    # Extract the (single) regex search results
    url = url[0]
    url_expression = url_expression[0]

    if url.startswith('/'):
        # Strip the leading url, so we can use it as a local path
        url = url[1:]

    path_font = path_base / Path(url)
    if not path_font.exists() and path_font.is_file():
        print(f'🔴 No font file found at url {url}. Skipping font inline conversion.')
        return line
    font_base64 = file_to_base64(path_font)
    url_inline = f'url(data:application/x-font-{format};charset=utf-8;base64,{font_base64})'

    return line.replace(url_expression, url_inline)


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

