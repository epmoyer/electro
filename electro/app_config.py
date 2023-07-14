"""Global app config."""

OUTPUT_FORMATS = ['static_site', 'single_file']

CONFIG = {
    'version': '1.5.0',
    'app_name': 'electro',
    'enable_debug_logging': True,
    'project_filename': 'electro.json',

    # ----------------------
    # Set at runtime
    # ----------------------
    'project_config': None,
    'path_project_directory': None,
    'path_output_directory': None,
    'path_theme_directory': None,
    'enable_newline_to_break': None,
    'output_format': None,
}
