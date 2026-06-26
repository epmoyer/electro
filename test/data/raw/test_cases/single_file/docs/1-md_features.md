# 1&nbsp;&nbsp;&nbsp;&nbsp;Markdown Features

## Test Execution Info

- Test Case: `single_file`
- Fields:
    - `app_name`: `@field{app_name}`
    - `app_version`: `@field{app_version}`
    - `datetime_now`: `@field{datetime_now}`

## Basic Formatting

This is text which occurs before the first internal heading.

Jackdaws love my big sphinx of quartz.

JACKDAWS LOVE MY BIG SPHINX OF QUARTZ. 

**Jackdaws love my big sphinx of quartz.**

**JACKDAWS LOVE MY BIG SPHINX OF QUARTZ.**

_Jackdaws love my big sphinx of quartz._

_JACKDAWS LOVE MY BIG SPHINX OF QUARTZ._

# Level 1 Heading
## Level 2 Heading
### Level 3 Heading
#### Level 4 Heading
##### Level 5 Heading
###### Level 6 Heading

## Link within document

In standard Markdown syntax: Link to [Line break test](#line-break-test)

In wiki-style syntax: Link to [[Line BrEak tEst ]]

## Link between documents
Link to [Overview](2-second.md#overview)

## Line break test
This line should not be
split from this line.

This line should not be
split from this line.

## Styles
**Bold**, _Italics Form 1_, *Italics Form 2*

## Tables
### Standard Markdown Syntax

This table is expressed in the "native" Markdown syntax.

| Column 1 | Column 2 | Column 3 |
| -------- | -------- | -------- |
| Data 1   | Data 2   | Data 3   |
| Data 4   | Data 5   | Data 6   |

### Embedded CSV syntax

Large and/or complex tables are often cumbersome to author and maintain in Markdown. The `@table` directive allows you to embed CSV data directly into your document.

Currently cells containing "n/a" (case insensitive) are automatically formatted with a gray background.  In the future this behavior will likely become configurable and support assignment of arbitrary background colors to arbitrary matching text content.

#### Simple CSV
@table{attachments/table_1.csv}

#### Complex CSV

@pragma{table_cell_bg_color_by_content:pass, #d0ffd0}
@pragma{table_cell_bg_color_by_content:fail, #ffd0d0}
@pragma{table_cell_bg_color_by_content:warning, #ffe8d0}
@pragma{table_cell_bg_color_by_content_partial:tbd, #ffff00}
@pragma{table_cell_bg_color_by_content:n/a, #c0c0c0}

With color pragmas applied:

@table{attachments/table_2.csv}

With color pragmas cleared:

@pragma{table_cell_bg_color_clear_all}

@table{attachments/table_2.csv}

## Link
This is a [Link to Google](https://www.google.com) in the middle of a line.

## Bullet list
- Top level
    - Indent 1
        - Indent 2
            - Indent 3
                - Indent 4
                    - Indent 5
                        - Indent 6
                            - Indent 7
                                - Indent 8
                                - Indent 8 again
                            - Indent 7 again
                            - Indent 7 yet again
- Back to top level
- Top level again

Paragraph Text
- Bullet list item 1, immediately following paragraph text
- Bullet list item 2

## Ordered list
1. This is the first
2. This is the second
3. This is the third

## Code

This is an `inline` code block

### Text
```text
This is a 
multiline
code block
```

### Python
```python
import json

def test_function(arg1, arg2):
    print(f'Arguments {arg1=}')
    print(json.dumps(arg2))
```

### Very Wide

#### Single line

```text
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Netus et malesuada fames
```

#### Multiple lines

```text
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Netus et malesuada fames.

ac turpis egestas maecenas pharetra convallis. Duis tristique sollicitudin nibh sit. Nunc sed augue lacus viverra vitae congue eu consequat. Dapibus ultrices.
```

#### fenced code block

```python
def foobar(foo):
    # ----------------------------------------------------------------------------------------------
    # Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Netus et malesuada fames
    # ac turpis egestas maecenas pharetra convallis. Duis tristique sollicitudin nibh sit. Nunc sed augue lacus viverra vitae congue eu consequat. Dapibus ultrices.
    # ----------------------------------------------------------------------------------------------
    if ('spam' in foo):
        if ('eggs' in foo):
            if ('pram' in foo):
                if ('walks' in foo):
                    if ('joke' in foo):
                        if ('parrot' in foo):
                            if ('cheese' in foo):
                                if ('brian' in foo):
                                    if ('spam' in foo):
                                        if ('eggs' in foo):
                                            if ('pram' in foo):
                                                if ('walks' in foo):
                                                    if ('joke' in foo):
                                                        if ('parrot' in foo):
                                                            if ('cheese' in foo):
                                                                if ('brian' in foo):
                                                                    print('Ha!')
    print(foo)
```

#### Fenced code block inside an unordered list

- Item 1
- Item 2
    - Item 2.1
        ```python
        import json

        def test_function(arg1, arg2):
            print(f'Arguments {arg1=}')
            print(json.dumps(arg2))
        ```
    - Item 2.2
- Item 3
- Item 4

## D2 Diagrams
See the [D2 Language](https://d2lang.com) website for details on D2 syntax.

### Hello World
```d2
direction: right
x -> y: hello world
```

### Terraform
This is the "Terraform Resources" example from the [D2 Examples](https://d2lang.com/examples/dagre/#) documentation.

```d2
vars: {
  d2-config: {
    layout-engine: elk
  }
}

*.style.font-size: 22
*.*.style.font-size: 22

title: |md
  # Terraform resources (v1.0.0)
| {near: top-center}

direction: right

project_connection: {
  style: {
    fill: "#C5C6C7"
    stroke: grey
  }
}

privatelink_endpoint: {tooltip: Datasource only}
group
group_partial_permissions
service_token
job: {
  style: {
    fill: "#ACE1AF"
    stroke: green
  }
}

conns: Connections (will be removed in the future,\nuse global_connection) {
  bigquery_connection
  fabric_connection
  connection

  bigquery_connection.style.fill: "#C5C6C7"
  fabric_connection.style.fill: "#C5C6C7"
  connection.style.fill: "#C5C6C7"
}
conns.style.fill: "#C5C6C7"

env_creds: Environment Credentials {
  grid-columns: 2
  athena_credential
  databricks_credential
  snowflake_credential
  bigquery_credential
  fabric_credential
  postgres_credential: {tooltip: Is used for Redshift as well}
  teradata_credential
}

service_token -- project: can scope to {
  style: {
    stroke-dash: 3
  }
}
group -- project
group_partial_permissions -- project
user_groups -- group
user_groups -- group_partial_permissions
project -- environment
project -- snowflake_semantic_layer_credential
job -- environment
job -- environment_variable_job_override
notification -- job
partial_notification -- job

webhook -- job: triggered by {
  style: {
    stroke-dash: 3
  }
}
environment -- global_connection
environment -- conns
global_connection -- privatelink_endpoint
global_connection -- oauth_configuration

environment -- env_creds
conns -- privatelink_endpoint
project -- project_repository
lineage_integration -- project
project_repository -- repository
environment -- environment_variable
environment -- partial_environment_variable
environment -- extended_attributes
environment -- semantic_layer_configuration
model_notifications -- environment

project -- project_connection {
  style: {
    stroke: "#C5C6C7"
  }
}
project_connection -- conns {
  style: {
    stroke: "#C5C6C7"
  }
}

(job -- *)[*].style.stroke: green
(* -- job)[*].style.stroke: green

account_level_settings: "Account level settings" {
  account_features
  ip_restrictions_rule
  license_map
  partial_license_map
}
account_level_settings.style.fill-pattern: dots
```

## Change Bars (Experimental)

**This content has a changebar on the single word "farmhouse" to test how it behaves during resizing etc.**

Whose woods these are I think I know. His house is in the village though; He will not see me stopping here To watch his woods fill up with snow. My little horse must think it queer To stop without a :change_bar_start farmhouse :change_bar_end near Between the woods and frozen lake The darkest evening of the year. He gives his harness bells a shake To ask if there is some mistake. The only other sound’s the sweep Of easy wind and downy flake. The woods are lovely, dark and deep, But I have promises to keep, And miles to go before I sleep, And miles to go before I sleep.


**This is a bunch of LOREM IPSUM text that will change where the change-bar below falls as the window is resized**

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec sodales mauris quis pulvinar vehicula. Pellentesque eleifend arcu eget sapien tincidunt dapibus. Vestibulum eget lorem blandit, viverra dui vel, dapibus ligula. Donec varius, nulla et molestie blandit, lectus metus dictum nunc, vel ornare diam quam nec ex. Vestibulum vel nunc mollis, tempor leo sit amet, molestie magna. Praesent lacinia eleifend metus nec pharetra. Proin a augue dictum, bibendum ex ac, volutpat odio. Vestibulum leo dui, varius id diam sed, ultrices vehicula justo. Suspendisse eget semper urna. Cras elementum in erat ut molestie. Vestibulum faucibus non felis eget pulvinar. Nullam rhoncus urna sed tellus vehicula tincidunt.

Nulla molestie pellentesque turpis, at tincidunt turpis cursus commodo. Nunc vitae consequat augue. Nunc ac diam a quam tincidunt consectetur et quis lorem. Proin porttitor porttitor mi, eu sagittis magna elementum et. Donec feugiat porta tortor a imperdiet. Nullam condimentum enim et accumsan maximus. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Nunc consectetur nec diam at egestas. Duis semper mauris vitae sem cursus scelerisque. Nulla gravida urna nisi, quis facilisis arcu pellentesque et. Aenean vel turpis odio. Maecenas eget neque suscipit, malesuada dolor et, aliquet neque. Sed nunc elit, suscipit non lacus eget, posuere bibendum sapien. Aliquam erat volutpat.

Pellentesque elementum est nec aliquam tempus. Cras dictum felis id nunc dictum, in cursus enim eleifend. Pellentesque suscipit auctor bibendum. Duis lacinia eleifend imperdiet. Nulla feugiat nunc imperdiet viverra sodales. Donec tempor finibus porttitor. Nunc volutpat fringilla justo vel auctor. Maecenas metus ligula, tristique a sem quis, mollis rhoncus leo.

Aliquam interdum, purus tristique tincidunt pellentesque, felis leo fermentum quam, quis rhoncus velit augue quis magna. Donec malesuada velit eu commodo efficitur. Curabitur facilisis pretium mauris, in mollis nisl commodo et. Proin velit libero, cursus eu aliquam id, molestie eu nulla. Morbi aliquet pretium urna at bibendum. Sed iaculis ultricies pretium. Donec non imperdiet eros. Aliquam pretium libero dolor, at bibendum nisi venenatis vitae. Aliquam ac commodo felis, sit amet maximus lectus. Aliquam sed suscipit orci. Integer suscipit, elit vitae tristique venenatis, justo tortor vulputate felis, id tempus erat nulla faucibus mauris. Maecenas id tempus sem. Phasellus tortor turpis, posuere ac tellus eget, ornare dapibus ipsum.

Donec elementum nisl scelerisque, elementum mi sit amet, egestas orci. In dignissim nisl finibus viverra sagittis. Quisque sed odio nisi. Maecenas vestibulum pharetra velit, id convallis ante iaculis non. Fusce consequat velit libero, eu pellentesque diam egestas commodo. Morbi mattis accumsan sapien, at blandit velit placerat sed. Etiam felis ipsum, venenatis a mauris et, condimentum ornare ante. Sed facilisis, ex vitae eleifend semper, tortor sapien lacinia arcu, non posuere dui enim eu diam. Vivamus nec velit eget orci convallis tincidunt nec vel neque. Nullam non metus libero. Sed vel odio sit amet elit pulvinar volutpat. In hac habitasse platea dictumst. Vestibulum dui dolor, ultrices sed massa ut, euismod consectetur erat. Etiam consectetur commodo ex, sed rhoncus orci porttitor vel. Vivamus ultricies, nunc sed porta scelerisque, leo sapien tincidunt metus, sit amet porta nunc leo sagittis ante. Nunc suscipit est in mauris semper venenatis.


:change_bar_start
This is a change bar test

- list item 1
- list item 2
    - list item 3 (indent)

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec sodales mauris quis pulvinar vehicula. Pellentesque eleifend arcu eget sapien tincidunt dapibus. Vestibulum eget lorem blandit, viverra dui vel, dapibus ligula. Donec varius, nulla et molestie blandit, lectus metus dictum nunc, vel ornare diam quam nec ex. Vestibulum vel nunc mollis, tempor leo sit amet, molestie magna. Praesent lacinia eleifend metus nec pharetra. Proin a augue dictum, bibendum ex ac, volutpat odio. Vestibulum leo dui, varius id diam sed, ultrices vehicula justo. Suspendisse eget semper urna. Cras elementum in erat ut molestie. Vestibulum faucibus non felis eget pulvinar. Nullam rhoncus urna sed tellus vehicula tincidunt.

## Heading inside a change bar
:change_bar_end

## Horizontal rule
Before the rule

---

After the rule

## Images
![](img/image1.png)

- Images inside bullet lists
- ![](img/image2.png)
    - ![](img/image3.png)

## Quote blocks
Plain text

> Block quote
> that is written on two lines

> This is a long blockquote written on a single line. Lorem ipsum dolor sit amet,consectetuer adipiscing elit. Aliquam hendrerit mi posuere lectus. Vestibulum enim wisi, viverra nec, fringilla in, laoreet vitae, risus.

Plain text

## As-run text

This is a line of regular text.

@/ This is as-run text /@

This is a line of text with @/ as-run text /@ inside of it.

This is a line of regular text.

@/

This is a block of as-run text.

This is a block of as-run text.

![](img/image2.png)

```
This is a fenced code block inside an as-run block.
```

/@

## Notice types

@block{info}

This is a test of a `info` notice. It has a tile above, an icon, and a place for text below. The text below can be any arbitrary markdown content.

```
This is preformatted text
inside a notice.
```

```python
# This is preformatted Python inside a notice.
import json

def test_function(arg1, arg2):
    print(f'Arguments {arg1=}')
    print(json.dumps(arg2))
```

@block{end}

@block{tip}

This is a test of a `tip` notice. It has a tile above, an icon, and a place for text below. The text below can be any arbitrary markdown content.

@block{end}

@block{note}
This is a test of a `note` notice. It has a tile above, an icon, and a place for text below. The text below can be any arbitrary markdown content.

- A bullet list
    - Item 1
    - Item 2
- Heding 2
    - Item 3
    - Item 4
@block{end}

@block{warning}
This is a test of a `warning` notice. It has a tile above, an icon, and a place for text below. The text below can be any arbitrary markdown content.
@block{end}

## Notice types (Legacy syntax)

{{% notice info %}}

This is a test of a `info` notice using the LEGACY syntax.

{{% /notice %}}

{{% notice tip %}}

This is a test of a `tip` notice using the LEGACY syntax.

{{% /notice %}}

{{% notice note %}}

This is a test of a `note` notice using the LEGACY syntax.

{{% /notice %}}

{{% notice warning %}}

This is a test of a `warning` notice using the LEGACY syntax.

{{% /notice %}}

@pragma{inject_heading_class:h-doc-info}
## Heading with injected class
This heading has injected class `h-doc-info`.

@pragma{inject_heading_class:}

@pragma{include_in_toc:false}
## Not included in Table of Contents
This section should NOT appear in the Table of Contents. 
@pragma{include_in_toc:true}

## Included in Table of Contents
This section should appear in the Table of Contents.

{{% table_of_contents %}}
