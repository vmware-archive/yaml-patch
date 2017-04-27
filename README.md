# yaml-patch

yaml-patch is a YAML-version of Evan Phoenix's
[json-patch](https://github.com/evanphx/json-patch), which is an implementation
of [JavaScript Object Notation (JSON) Patch](https://tools.ietf.org/html/rfc6902)
in Go.

## Usage

```
  // Given a source document as []byte
  doc := []byte(`---
foo: bar
baz:
  quux: grault
`)

  // And an array of operations as []byte
  ops := []byte(`---
- op: add
  path: /baz/waldo
  value: fred
`)

  // Decode the operations into a Patch
  patch, err := yamlpatch.DecodePatch(ops)
  if err != nil {
    log.Fatalf("decoding patch failed: %s", err)
  }

  // Apply the patch to the document
  bs, err := patch.Apply(doc)
  if err != nil {
    log.Fatalf("applying patch failed: %s", err)
  }

  // Gander in awe at your modified document
  fmt.Println(string(bs))
  // baz:
  //   quux: grault
  //   waldo: fred
  // foo: bar
```
