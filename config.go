package jsoncfg

import (
    "encoding/json"
    "fmt"
    "reflect"
)

// Slice wrapper is needed to get a reflect.Type for interface{}
var interfaceSlice []interface{}
var interfaceType = reflect.TypeOf(interfaceSlice).Elem()

type Parser struct {
    def reflect.Value   // Default value
    sub []reflect.Value // Sub default values
    typ reflect.Type    // Struct type that has its types converted to interfaces
    vf  map[uintptr] reflect.Value // Validator functions
}

func NewParser(pstr interface{}) (*Parser, error) {

    // Ensure pstr is pointer to struct
    // pstr needs to be a pointer in order to be an addressable value
    if isPtrToStruct(pstr) == false {
        return nil, fmt.Errorf("The given parameter is not a pointer to struct")
    }

    // Struct to interface struct
    def := reflect.ValueOf(pstr).Elem()
    typ := fieldsToInterface(def.Type())
    
    // Return
    return &Parser{
        def: def,
        typ: typ,
        sub: make([]reflect.Value, 0),
        vf: make(map[uintptr] reflect.Value),
    }, nil

}

// Make fields whose zero value is not nil into interfaces
func fieldsToInterface(typ reflect.Type) reflect.Type {

    nf     := typ.NumField()
    fields := make([]reflect.StructField, 0)
    verify := func(rhs reflect.Type) {
        if typ == rhs {panic("Nesting of same type is not allowed")}
    }
    
    for i := 0; i < nf; i++ {

        field := typ.Field(i)

        // Check if exported
        first := field.Name[0]
        if first < 'A' || first > 'Z' {
            continue
        }

        kind := field.Type.Kind()
        switch kind {
        case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
            reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
            reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
            reflect.String:
            field.Type = interfaceType
        case reflect.Slice, reflect.Map:

            // Check for struct
            ftel := field.Type.Elem()
            if ftel.Kind() != reflect.Struct {break}

            verify(ftel)

            switch kind {
            case reflect.Slice:
                field.Type = reflect.SliceOf(fieldsToInterface(ftel))
            case reflect.Map:
                field.Type = reflect.MapOf(
                    field.Type.Key(), fieldsToInterface(ftel),
                )
            }

        case reflect.Struct:
            
            // Recursive
            verify(field.Type)
            field.Type = fieldsToInterface(field.Type)

        default:
        }

        fields = append(fields, field)

    }

    return reflect.StructOf(fields)

}

func isPtrToStruct(pstr interface{}) bool {
    rv := reflect.ValueOf(pstr)
    return rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct
}

// PARSER ---

func(p *Parser) Parse(data []byte, pstr interface{}) (err error) {

    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("%v", r)
        }
    }()

    if isPtrToStruct(pstr) == false {
        return fmt.Errorf("The given parameter is not a pointer to struct")
    }

    // Ensure same type as default config
    rv := reflect.ValueOf(pstr)
    el := rv.Elem()
    if el.Type() != p.def.Type() {
        return fmt.Errorf("The given struct is not the same type as the default configuration")
    }

    // Make new interface-fied struct
    // and unmarshal json content into it
    pa  := reflect.New(p.typ)
    a   := pa.Elem()
    err  = json.Unmarshal(data, pa.Interface())
    if err != nil {
        return err
    }

    // Destination struct which is the same type as the default struct
    pb := reflect.New(p.def.Type())
    b  := pb.Elem()

    // Deep fill nil
    p.deepFillNil(p.def, a, b)

    // Assign
    el.Set(b)

    return nil

}

// Put a's contents into b filling nils with default values as defined in def
// a must be struct that has contents unmarshaled by json package
func(p *Parser) deepFillNil(def, a, b reflect.Value) { // a => b

    for i := 0; i < def.NumField(); i++ { // def is standards

        name := def.Type().Field(i).Name

        // Check if exported
        first := name[0]
        if first < 'A' || first > 'Z' {
            continue
        }

        dv := def.Field(i) // Default value for current member
        av := a.FieldByName(name) // as av has fewer fields, find fields by name
        bv := b.Field(i)

        if bv.Type().Kind() == reflect.Struct {

            // If it is a struct member, do recursive call
            p.deepFillNil(dv, av, bv)

        } else {

            // Non-struct

            if av.IsNil() { // nil interface value

                // Nil value means that the member was not found in a
                // Therefore, put default value
                bv.Set(dv)

            } else { // av has value
                
                // Convert interfaces
                switch bv.Type().Kind() {

                case reflect.Bool:
            
                    v := av.Interface().(bool)
                    bv.SetBool(v)
            
                case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
                    reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
                    reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
            
                    // All numbers in a is float64 as encoding/json decoded it
                    rf64 := reflect.ValueOf(av.Interface().(float64))
                    bv.Set(rf64.Convert(bv.Type()))
            
                case reflect.String:
            
                    s := av.Interface().(string)
                    bv.SetString(s)

                case reflect.Slice, reflect.Map:

                    bvElTyp := bv.Type().Elem()

                    if bvElTyp.Kind() == reflect.Struct {

                        // Check for sub defaults
                        // Zero value for bv elem type is fallback in case of no sub default
                        subDv := reflect.New(bvElTyp).Elem()
                        for _, each := range p.sub {
                            // Compare element type
                            if bvElTyp == each.Type() {
                                subDv = each
                                break
                            }
                        }

                        // Deep copy each element
                        switch bv.Type().Kind() {
                        case reflect.Slice:
                            bv.Set(reflect.MakeSlice(bv.Type(), 0, 0))
                            for k := 0; k < av.Len(); k++ {
                                subAv := av.Index(k)
                                subBv := reflect.New(bvElTyp).Elem()
                                p.deepFillNil(subDv, subAv, subBv)
                                bv.Set(reflect.Append(bv, subBv))
                            }
                        case reflect.Map:
                            bv.Set(reflect.MakeMap(bv.Type()))
                            keys := av.MapKeys()
                            for _, key := range keys {
                                subAv := av.MapIndex(key)
                                subBv := reflect.New(bvElTyp).Elem()
                                p.deepFillNil(subDv, subAv, subBv)
                                bv.SetMapIndex(key, subBv)
                            }
                        }

                    } else {
                        bv.Set(av)
                    }

                default:
                    bv.Set(av)
                }

            }

            // Find validator by its default value's address
            rvf, ok := p.vf[dv.Addr().Pointer()]
            if ok {
                ins := make([]reflect.Value, 1)
                bvt := bv.Type()
                switch rvf.Type().In(0) {
                case reflect.PtrTo(bvt): ins[0] = bv.Addr()
                case bvt:                ins[0] = bv
                }

                out   := rvf.Call(ins)[0]
                valid := out.Bool()
                if !valid {
                    panic(fmt.Errorf(
                        "%s.%s has an invalid value of %v",
                        b.Type().Name(), b.Type().Field(i).Name, bv,
                    ))
                }
            }

        }
    }

}


func(p *Parser) SetValidator(ptr, vf interface{}) error {

    // SetValidator sets a validator function for the entries
    // that correspond to the pointer that is part of the default value
    // The pointer must point to a member of the default value or the sub default values

    rptr := reflect.ValueOf(ptr)
    rel  := rptr.Elem()
    rvf  := reflect.ValueOf(vf)

    // Ensure function is func(type) bool or func(*type) bool
    if rvf.Type().NumIn() != 1 {
        return fmt.Errorf("Given function has invalid parameter count")
    }
    switch rvf.Type().In(0) {
    case rel.Type(), rptr.Type():
    default:
        return fmt.Errorf(
            "Wrong parameter type, %v, for validator function for %v",
            rvf.Type().In(0), rel.Type(),
        )
    }
    if rvf.Type().NumOut() != 1 || rvf.Type().Out(0).Kind() != reflect.Bool {
        return fmt.Errorf("Wrong return type for validator function")
    }

    // Assign
    p.vf[rptr.Pointer()] = rvf

    return nil

}

// Sub default definitions
func(p *Parser) SetSubDefault(pstr interface{}) error {

    // Add parsers for structs inside array, map, or slice

    // Ensure pstr is struct
    if isPtrToStruct(pstr) == false {
        return fmt.Errorf("The given parameters is not a pointer to struct")
    }

    // Struct to interface struct
    sub := reflect.ValueOf(pstr).Elem()

    // Prepend so as to take precednce over the already added
    p.sub = append([]reflect.Value{sub}, p.sub...)

    return nil

}