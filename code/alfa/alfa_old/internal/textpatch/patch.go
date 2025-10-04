package textpatch

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "strings"
)

type PatchOp struct {
    Line    int      `json:"line"`
    Type    string   `json:"type"`
    Content []string `json:"content,omitempty"`
}

// parsePatchOps decodes patch JSON string into PatchOp slice
func parsePatchOps(patchJson string) ([]PatchOp, error) {
    var ops []PatchOp
    err := json.Unmarshal([]byte(patchJson), &ops)
    if err != nil {
        return nil, err
    }
    return ops, nil
}

// patchLines applies the Codex patch format to a slice of lines.
func PatchLines(inputLines []string, patchJson string) ([]string, error) {
    ops, err := parsePatchOps(patchJson)
    if err != nil {
        return nil, err
    }
    // Defensive copy
    lines := make([]string, len(inputLines))
    copy(lines, inputLines)

    // Ensure ops are sorted by .Line
    for i := 1; i < len(ops); i++ {
        if ops[i].Line < ops[i-1].Line {
            return nil, errors.New("patch operations must be sorted by line")
        }
    }

    // Always use CURRENT pos for CURRENT lines
    for _, op := range ops {
        pos := op.Line
        switch op.Type {
        case "insert":
            if pos < 0 || pos > len(lines) {
                return nil, errors.New("insert position out of range")
            }
            lines = append(lines[:pos], append(op.Content, lines[pos:]...)...)
        case "delete":
            if pos < 0 || pos >= len(lines) {
                return nil, errors.New("delete position out of range")
            }
            lines = append(lines[:pos], lines[pos+1:]...)
        case "replace":
            if pos < 0 || pos >= len(lines) {
                return nil, errors.New("replace position out of range")
            }
            lines = append(lines[:pos], append(op.Content, lines[pos+1:]...)...)
        default:
            return nil, errors.New("invalid patch type: " + op.Type)
        }
    }
    return lines, nil
}

// PatchText applies a patch to a string and returns the new string.
func PatchText(input string, patchJson string) (string, error) {
    lines := strings.Split(input, "\n")
    patched, err := PatchLines(lines, patchJson)
    if err != nil {
        return "", err
    }
    return strings.Join(patched, "\n"), nil
}

// PatchFile reads a file and writes the result to the same file (in-place)
func PatchFile(inputFile string, patchJson string, outputFiles ...string) error {
    data, err := ioutil.ReadFile(inputFile)
    if err != nil {
        return err
    }
    lines := strings.Split(string(data), "\n")
    patched, err := PatchLines(lines, patchJson)
    if err != nil {
        return err
    }
    outdata := strings.Join(patched, "\n")
    outFile := inputFile
    if len(outputFiles) > 0 && outputFiles[0] != "" {
        outFile = outputFiles[0]
    }
    return ioutil.WriteFile(outFile, []byte(outdata), 0644)
}
