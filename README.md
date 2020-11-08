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

### Fallback for Structs in Slices or Maps

Package jsoncfg can put default values for the structs in slices or maps.

```go

type SubCfg struct {
    B1 int `json:"b1"`
}

type MyCfg struct {
    A []SubCfg `json:"a"`
}

var data = []byte(`{
    "a": [
        {}
    ]
}`)

var DefaultSubCfg = SubCfg{
    B1: 50,
}

var DefaultMyCfg = MyCfg{
    A: []SubCfg{},
}

func do() {
    var cfg MyCfg
    parser, _ := jsoncfg.NewParser(DefaultMyCfg)
    parser.ChildDefaults{ // Register child defaults
        &DefaultSubCfg,
    }
    _ = parser.Parse(data, &cfg)
    fmt.Println(cfg)

    // MyCfg{
    //   A: []SubCfg{
    //     SubCfg{
    //       B1: 50,
    //     },
    //   },
    // }
}

```

### Validators

You can use validator functions to verify configurations.

```go

type MyCfg struct {
    A int `json:"a"`
    B int `json:"b"`
}

var DefaultMyCfg = MyCfg{
    A: 1, B: 2,
}

var data = []byte(`{
    "a": -1,
    "b": 22
}`)

func do() {

    var cfg MyCfg
    parser, _ := jsoncfg.NewParser(DefaultMyCfg)
    parser.Validator(&DefaultMyCfg.A, func(v int) bool {
        return i > 0
    })

    _ = parser.Parse(data, &cfg)
    // Error: A has an invalid value of -1

    parser.Validator(&DefaultMyCfg.A, func(v int) bool {
        return true
    })
    parser.Validator(&DefaultMyCfg.B, func(pv *int) bool) { // Notice the pointer
        if *p == 22 {
            *p = 3 // You can change the value in the validator
        }
        return true
    }

    _ = parser.Parse(data, &cfg)
    // OK

    fmt.Println(cfg)
    // MyCfg{
    //   A: -1, B: 3,
    // }

}

```