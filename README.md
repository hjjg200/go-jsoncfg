# go-jsoncfg

```go
import "github.com/hjjg200/go-jsoncfg"
```

Package jsoncfg provides utility functions for configuration files in json encoding.

This package is not at the final version at the moment; function names are liable to change; forking is recommended.

## Example

### Fallback for Missing Members

After adding new members to configuration struct, package json will decode the old json config file and leave the newly added members set as zero values. But the following example will fill the default values for the newly added values.

```go

var oldCfg = []byte(`{
    "a": 1
}`) // version 1

type MyCfg struct { // version 2
    A int `json:"a"`
    B int `json:"b"`
}

var DefaultMyCfg = MyCfg{
    A: 0, B: 100,
}

func do() {
    var cfg MyCfg
    parser, _ := jsoncfg.NewParser(DefaultMyCfg)
    _ = parser.Parse(oldCfg, &cfg)
    fmt.Println(cfg)
    
    // MyCfg{
    //   A: 1, B: 100,
    // }
}

```


