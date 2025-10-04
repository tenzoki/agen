# Codex Patch Format

This document defines the Codex patch format, which is a structured JSON format for applying line-level changes to text files.

## Patch Object Structure

### Formal description

```
[
  {
    "line": <int>,          // Line number (0-based index) at which the operation applies
    "type": "insert" | "delete" | "replace",
    "content": [<string>]   // Required for insert/replace; omitted for delete
  }
]
```

Each patch is a JSON object with the following fields:

- `line` (int): The 0-based line number in the target file where the patch applies.
- `type` (string): One of `"insert"`, `"delete"`, or `"replace"`.
- `content` (array of strings): The lines to insert or replace (not used for delete).

## Supported Patch Types

### Insert

```json
{
  "line": 3,
  "type": "insert",
  "content": ["inserted line 1", "inserted line 2"]
}
```

Inserts the given lines *before* line 3.

### Delete

```json
{
  "line": 5,
  "type": "delete"
}
```

Deletes the line at index 5.

### Replace ✅

```json
{
  "line": 10,
  "type": "replace",
  "content": [
    "new line 1",
    "new line 2"
  ]
}
```

Replaces line 10 with the provided lines. This may expand to multiple lines or contract to a single one. It behaves like a `delete` followed by an `insert` at the same position.

## Patch Application Rules

- Patches must be applied in ascending order of `line`.
- Deletions affect the line numbering of subsequent patches.
- When applying:
  - For `insert`, insert content before the line number.
  - For `delete`, remove the specified line.
  - For `replace`, remove the specified line and insert the new content in its place.

## Example

Original file:

```
0: line one
1: line two
2: line three
3: line four
```

Patch set:

```json
[
  {"line": 1, "type": "replace", "content": ["new two", "new three"]},
  {"line": 3, "type": "delete"},
  {"line": 3, "type": "insert", "content": ["inserted line"]}
]
```

Result:

```
0: line one
1: new two
2: new three
3: inserted line
```

Note that line numbers shift as changes are applied in order.