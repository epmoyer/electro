from electro.faults import FAULTS

SNIPPET_NOTICE_START_TEMPLATE = (
    '<div class="notices [[notice_type]]">'
    '<div class="label"><i class="fa [[icon]]"></i>[[title]]</div>'
)
SNIPPET_NOTICE_END = '</div>'

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
    return (
        SNIPPET_NOTICE_START_TEMPLATE.replace('[[notice_type]]', notice_type)
        .replace('[[icon]]', icon)
        .replace('[[title]]', notice_type.capitalize())
    )

