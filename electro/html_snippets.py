from electro.faults import FAULTS

SNIPPET_NOTICE_START_TEMPLATE = (
    '<div class="notices [[notice_type]]">'
    '<div class="label"><i class="fa [[icon]]"></i>Info</div>'
)
SNIPPET_NOTICE_END_TEMPLATE = '</div>'

NOTICE_ICONS = {
    'note': 'fa-exclamation-circle',
    'info': 'fa-info-circle',
    'tip': 'fa-wrench',
    'warning': 'fa-warning',
}


def build_snippet_notice_start(notice_type):
    icon = NOTICE_ICONS.get(notice_type)
    if notice_type is None:
        FAULTS.error(f'Unrecognized notice_type: {notice_type}')
        return ''
    return SNIPPET_NOTICE_START_TEMPLATE.replace('[[notice_type', notice_type).replace(
        '[[icon]]', icon
    )


def build_snippet_notice_end(notice_type):
    return '' if notice_type not in NOTICE_ICONS else SNIPPET_NOTICE_END_TEMPLATE

