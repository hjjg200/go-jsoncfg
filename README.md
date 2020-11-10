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

oldJson := []byte(`{"A": 10}`)
type Foo struct{A, B int}  // Version 2: B is added
defFoo := Foo{A: 0, B: 22} // Default foo value
parser, _ := NewParser(&defFoo)

var myFoo Foo // Empty foo value
parser.Parse(oldJson, &myFoo)

fmt.Println(myFoo)
// {10 22}

```

### Fallback for Structs in Slices or Maps

Package jsoncfg can put default values for the structs in slices or maps.

```go
type Bar struct{Name string} // Sub member
type Foo struct{Bars []Bar}  // Slice of struct

var defFoo Foo
var defBar = Bar{Name: "it's bar"}
parser, _ := NewParser(&defFoo)
parser.SetSubDefault(&defBar) // Set default value for Bar

// Parse
var myFoo Foo
data := []byte(`{
    "Bars": [{}, {}]
}`) // Two empty bars
parser.Parse(data, &myFoo)

fmt.Println(myFoo)
// {[{it's bar} {it's bar}]}
```

### Validators

You can use validator functions to verify configurations.

```go
type Foo struct{Odd, IAM22 int}
var defFoo = Foo{Odd: 1, IAM22: 22}
parser, _ := NewParser(&defFoo)

parser.SetValidator(&defFoo.Odd, func(v int) bool {
    return v % 2 == 1
})
parser.SetValidator(&defFoo.IAM22, func(pv *int) bool { // Notice it is a pointer
    *pv = 22 // You can change the value inside a validator
    return true
})

// Parse
var myFoo Foo

// Wrong data
abomination := []byte(`{
    "Odd": 2
}`)
fmt.Println(parser.Parse(abomination, &myFoo))
// ERROR: Foo.Odd has an invalid value of 2

data := []byte(`{
    "Odd": 3, "IAM22": -1
}`)
parser.Parse(data, &myFoo)
fmt.Println(myFoo)
// {3 22}
```