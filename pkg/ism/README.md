# Smooth Streaming parser library

[![Reference](https://godoc.org/github.com/davimdo/gott/pkg/ism?status.svg)](https://godoc.org/github.com/davimdo/gott/pkg/ism)

Simple library to parse Smooth Streaming ISM files.

Smooth Streaming specification: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-sstr


## Example

```go
import "github.com/davimdo/gott/pkg/ism"

ssm, err := smooth.Unmarshal(ism)
if err != nil {
    t.Fatal(err)
}
fmt.Println(ssm)

b, err := ssm.Marshal()
if err != nil {
    t.Fatal(err)
}
fmt.Println(string(b))
```