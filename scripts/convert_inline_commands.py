#!/usr/bin/env python3
"""
Conservative converter: replace inline backtick CLI examples with fenced code blocks.
Rules:
 - Only convert inline code spans that contain whitespace (likely commands).
 - Skip spans that contain '/' or end with '.md' (these are file links) to preserve docs links.
 - Skip content inside existing fenced code blocks.
 - For table cells: replace cell content with `(see example below)` and insert code blocks after the table.
 - For plain paragraphs or list items: remove the inline span from the line and insert a fenced code block after the line.
 - Avoid inserting duplicate code blocks within the same file.

Usage: run from repo root. It will edit files in place and print a summary.
"""

import os
import re
from pathlib import Path

ROOT = Path('docs')
BACKTICK_RE = re.compile(r"`([^`]*\s[^`]*)`")
FILELIKE_RE = re.compile(r"/|\.md$")

files_changed = []

for path in ROOT.rglob('*.md'):
    with path.open('r', encoding='utf-8') as f:
        lines = f.readlines()

    out_lines = []
    in_fence = False
    fence_delim = None
    i = 0
    pending_table_codeblocks = []
    file_codeblocks = []  # avoid duplicates
    changed = False

    while i < len(lines):
        line = lines[i]
        # detect fenced code block start/end
        m = re.match(r"^(?P<delim>`{3,}|~{3,})(?P<lang>.*)$", line)
        if m:
            delim = m.group('delim')
            if not in_fence:
                in_fence = True
                fence_delim = delim
            else:
                if delim == fence_delim:
                    in_fence = False
                    fence_delim = None
            out_lines.append(line)
            i += 1
            continue

        if in_fence:
            out_lines.append(line)
            i += 1
            continue

        # detect table start: a line with '|' and next line with ---|---
        if '|' in line:
            # naive table handling: collect table block
            table_start = i
            table_lines = []
            while i < len(lines) and (lines[i].strip() != '' and '|' in lines[i]):
                table_lines.append(lines[i])
                i += 1
            # process table rows
            new_table_lines = []
            table_changed = False
            table_codeblocks = []
            for row in table_lines:
                # find backtick spans in the row
                table_row_codes = []
                def repl_table(m):
                    inner = m.group(1).strip()
                    if FILELIKE_RE.search(inner):
                        return m.group(0)  # keep
                    # it's a command; replace cell with marker
                    code = inner
                    if code not in table_row_codes:
                        table_row_codes.append(code)
                    return '(see example below)'

                new_row = BACKTICK_RE.sub(repl_table, row)
                new_table_lines.append(new_row)
            out_lines.extend(new_table_lines)
            # collect any codes from the table rows
            row_codes = []
            for r in table_lines:
                for m in BACKTICK_RE.finditer(r):
                    inner = m.group(1).strip()
                    if FILELIKE_RE.search(inner):
                        continue
                    if inner not in row_codes:
                        row_codes.append(inner)
            if row_codes:
                changed = True
                for cmd in row_codes:
                    if cmd in file_codeblocks:
                        continue
                    file_codeblocks.append(cmd)
                    pending_table_codeblocks.append(cmd)
            # if next line is blank, keep it
            if i < len(lines) and lines[i].strip() == '':
                out_lines.append(lines[i])
                i += 1
            # insert pending table blocks now
            if pending_table_codeblocks:
                for cmd in pending_table_codeblocks:
                    out_lines.append('\n')
                    out_lines.append('```bash\n')
                    out_lines.append(cmd + '\n')
                    out_lines.append('```\n')
                pending_table_codeblocks = []
            continue

        # normal line processing
        found = []
        def repl(m):
            inner = m.group(1).strip()
            if FILELIKE_RE.search(inner):
                return m.group(0)
            # command to extract
            found.append(inner)
            return ''  # remove inline code from line

        new_line = BACKTICK_RE.sub(repl, line)
        if found:
            changed = True
            # append the modified line
            out_lines.append(new_line)
            # add code blocks for unique commands not yet in file
            for cmd in found:
                if cmd in file_codeblocks:
                    continue
                file_codeblocks.append(cmd)
                out_lines.append('\n')
                out_lines.append('```bash\n')
                out_lines.append(cmd + '\n')
                out_lines.append('```\n')
        else:
            out_lines.append(line)
        i += 1

    if changed:
        with path.open('w', encoding='utf-8') as f:
            f.writelines(out_lines)
        files_changed.append(str(path))

# summary
print('Files changed: %d' % len(files_changed))
for p in files_changed:
    print(' -', p)

if not files_changed:
    print('No changes made.')

