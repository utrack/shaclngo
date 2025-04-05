package rdft

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/deiu/rdf2go"
)

// Unmarshaller converts RDF triples to Go structs
type Unmarshaller struct {
	graph *rdf2go.Graph
	// StrictMode when true will return an error if a subject has predicates
	// that don't match any field in the target struct
	StrictMode bool
}

// UnmarshallerOption is a function that configures an Unmarshaller
type UnmarshallerOption func(*Unmarshaller)

// WithStrictMode enables strict mode which returns an error when a subject has
// predicates that don't match any field in the target struct
func WithStrictMode() UnmarshallerOption {
	return func(u *Unmarshaller) {
		u.StrictMode = true
	}
}

// NewUnmarshaller creates a new unmarshaller with the given RDF graph
func NewUnmarshaller(graph *rdf2go.Graph, opts ...UnmarshallerOption) *Unmarshaller {
	u := &Unmarshaller{
		graph: graph,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(u)
	}
	
	return u
}

// Unmarshal unmarshals RDF data into the given Go struct
func (u *Unmarshaller) Unmarshal(subject string, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("unmarshal target must be a non-nil pointer")
	}

	elem := val.Elem()
	typ := elem.Type()

	// Check if the type implements RDFUnmarshaler
	if u.implementsRDFUnmarshaler(val) {
		// Get all triples with the given subject
		var triples []*rdf2go.Triple
		
		// Handle different subject types (resource or blank node)
		if strings.HasPrefix(subject, "_:") {
			// This is a blank node
			blankNodeID := strings.TrimPrefix(subject, "_:")
			triples = u.graph.All(rdf2go.NewBlankNode(blankNodeID), nil, nil)
		} else {
			// This is a resource
			triples = u.graph.All(rdf2go.NewResource(subject), nil, nil)
		}
		
		// Convert to Values
		values := make([]Value, 0, len(triples))
		for _, triple := range triples {
			values = append(values, FromRDF2GoTerm(triple.Object))
		}
		
		// Call UnmarshalRDF
		return val.Interface().(RDFUnmarshaler).UnmarshalRDF(values)
	}

	// Get all triples for this subject
	var subjectTriples []*rdf2go.Triple
	
	// Handle different subject types (resource or blank node)
	if strings.HasPrefix(subject, "_:") {
		// This is a blank node
		blankNodeID := strings.TrimPrefix(subject, "_:")
		subjectTriples = u.graph.All(rdf2go.NewBlankNode(blankNodeID), nil, nil)
		
		// Debug information
		fmt.Printf("Debug: Unmarshalling blank node %s, found %d triples\n", 
			subject, len(subjectTriples))
	} else {
		// This is a resource
		subjectTriples = u.graph.All(rdf2go.NewResource(subject), nil, nil)
	}
	
	// If in strict mode, collect all predicates from the struct
	knownPredicates := make(map[string]bool)
	if u.StrictMode {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			predicate := field.Tag.Get("rdf")
			if predicate != "" {
				knownPredicates[predicate] = true
			}
		}
	}

	// Process struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := elem.Field(i)

		// Skip unexported fields
		if !fieldVal.CanSet() {
			continue
		}

		// Get predicate from tag
		predicate := field.Tag.Get("rdf")
		if predicate == "" {
			continue
		}

		// Check if field should be treated as localized
		// Either by explicit tag or by field type
		rdfType := field.Tag.Get("rdf-type")
		fieldType := fieldVal.Type()
		isLocalizedType := fieldType == reflect.TypeOf(LocalizedString{}) || 
			fieldType == reflect.TypeOf(&LocalizedString{}) || 
			fieldType == reflect.TypeOf(LocalizedText{}) || 
			fieldType == reflect.TypeOf(&LocalizedText{})
		
		if rdfType == "localized" || isLocalizedType {
			if err := u.unmarshalLocalizedField(subject, predicate, fieldVal); err != nil {
				return err
			}
			continue
		}

		// Handle regular fields
		if err := u.unmarshalField(subject, predicate, fieldVal); err != nil {
			return err
		}
	}

	// In strict mode, check for unknown predicates
	if u.StrictMode {
		var unknownPredicates []string
		for _, triple := range subjectTriples {
			predURI := triple.Predicate.(*rdf2go.Resource).URI
			if !knownPredicates[predURI] {
				unknownPredicates = append(unknownPredicates, predURI)
			}
		}
		
		if len(unknownPredicates) > 0 {
			return fmt.Errorf("unknown predicates for subject %s: %v", subject, unknownPredicates)
		}
	}

	return nil
}

// unmarshalField unmarshals a single field
func (u *Unmarshaller) unmarshalField(subject, predicate string, field reflect.Value) error {
	// Handle different subject types (resource or blank node)
	var subjectTerm rdf2go.Term
	if strings.HasPrefix(subject, "_:") {
		// This is a blank node
		blankNodeID := strings.TrimPrefix(subject, "_:")
		subjectTerm = rdf2go.NewBlankNode(blankNodeID)
	} else {
		// This is a resource
		subjectTerm = rdf2go.NewResource(subject)
	}
	
	triples := u.graph.All(
		subjectTerm,
		rdf2go.NewResource(predicate),
		nil,
	)

	if len(triples) == 0 {
		return nil // No value to set
	}

	// Get first triple (for single-valued fields)
	triple := triples[0]
	objValue := FromRDF2GoTerm(triple.Object)

	// Handle different field types
	switch field.Kind() {
	case reflect.String:
		// Check if the object is a resource - if so, this is a type mismatch
		if _, isResource := objValue.(*Resource); isResource {
			return fmt.Errorf("type mismatch: cannot unmarshal Resource into string field, use rdft.Resource instead")
		}
		field.SetString(objValue.RawValue())
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(objValue.RawValue(), 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to int: %w", objValue.RawValue(), err)
		}
		field.SetInt(i)
		
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(objValue.RawValue(), 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to uint: %w", objValue.RawValue(), err)
		}
		field.SetUint(i)
		
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(objValue.RawValue(), 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to float: %w", objValue.RawValue(), err)
		}
		field.SetFloat(f)
		
	case reflect.Bool:
		b, err := strconv.ParseBool(objValue.RawValue())
		if err != nil {
			return fmt.Errorf("cannot convert %s to bool: %w", objValue.RawValue(), err)
		}
		field.SetBool(b)
		
	case reflect.Map:
		// Handle map types
		if field.Type().Key().Kind() == reflect.String && field.Type().Elem().Kind() == reflect.String {
			// Initialize the map if it's nil
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
			
			// Handle different types of objects
			switch obj := triple.Object.(type) {
			case *rdf2go.Resource:
				// For resources, get all predicates and their values
				resourceTriples := u.graph.All(
					obj,
					nil,
					nil,
				)
				
				// Add each predicate-object pair to the map
				for _, t := range resourceTriples {
					predURI := t.Predicate.(*rdf2go.Resource).URI
					objVal := FromRDF2GoTerm(t.Object)
					
					// Use the local part of the predicate URI as the key
					parts := strings.Split(predURI, "#")
					key := predURI
					if len(parts) > 1 {
						key = parts[1]
					} else {
						parts = strings.Split(predURI, "/")
						if len(parts) > 0 {
							key = parts[len(parts)-1]
						}
					}
					
					field.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(objVal.RawValue()))
				}
				
			case *rdf2go.BlankNode:
				// For blank nodes, get all predicates and their values
				resourceTriples := u.graph.All(
					obj,
					nil,
					nil,
				)
				
				// Add each predicate-object pair to the map
				for _, t := range resourceTriples {
					predURI := t.Predicate.(*rdf2go.Resource).URI
					objVal := FromRDF2GoTerm(t.Object)
					
					// Use the local part of the predicate URI as the key
					parts := strings.Split(predURI, "#")
					key := predURI
					if len(parts) > 1 {
						key = parts[1]
					} else {
						parts = strings.Split(predURI, "/")
						if len(parts) > 0 {
							key = parts[len(parts)-1]
						}
					}
					
					field.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(objVal.RawValue()))
				}
				
			default:
				// For other types, just store the raw value with a default key
				field.SetMapIndex(reflect.ValueOf("value"), reflect.ValueOf(objValue.RawValue()))
			}
			
			return nil
		}
		return fmt.Errorf("unsupported map type: %s", field.Type().String())
		
	case reflect.Struct:
		// Handle special types like time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := time.Parse(time.RFC3339, objValue.RawValue())
			if err != nil {
				return fmt.Errorf("cannot convert %s to time.Time: %w", objValue.RawValue(), err)
			}
			field.Set(reflect.ValueOf(t))
		} else {
			// For other structs, handle different object types differently
			switch obj := triple.Object.(type) {
			case *rdf2go.Resource:
				// For resources, unmarshal directly
				newVal := reflect.New(field.Type())
				if err := u.Unmarshal(obj.URI, newVal.Interface()); err != nil {
					return err
				}
				field.Set(newVal.Elem())
			
			case *rdf2go.BlankNode:
				// For blank nodes, unmarshal the blank node as a subject
				newVal := reflect.New(field.Type())
				if err := u.Unmarshal(obj.String(), newVal.Interface()); err != nil {
					return err
				}
				field.Set(newVal.Elem())
			
			default:
				// For other types, try to unmarshal using the raw value
				newVal := reflect.New(field.Type())
				if err := u.Unmarshal(objValue.RawValue(), newVal.Interface()); err != nil {
					return err
				}
				field.Set(newVal.Elem())
			}
		}
		
	case reflect.Ptr:
		// Create a new instance of the pointed-to type
		ptrType := field.Type().Elem()
		newVal := reflect.New(ptrType)
		
		// Handle different pointed-to types
		switch ptrType.Kind() {
		case reflect.String:
			newVal.Elem().SetString(objValue.RawValue())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(objValue.RawValue(), 10, 64)
			if err != nil {
				return fmt.Errorf("cannot convert %s to int: %w", objValue.RawValue(), err)
			}
			newVal.Elem().SetInt(i)
		case reflect.Struct:
			if ptrType == reflect.TypeOf(Resource{}) {
				if res, ok := objValue.(*Resource); ok {
					newVal.Elem().Set(reflect.ValueOf(*res))
				} else {
					newVal.Elem().Set(reflect.ValueOf(Resource{URI: objValue.RawValue()}))
				}
			} else if ptrType == reflect.TypeOf(Literal{}) {
				if lit, ok := objValue.(*Literal); ok {
					newVal.Elem().Set(reflect.ValueOf(*lit))
				} else {
					newVal.Elem().Set(reflect.ValueOf(Literal{Value: objValue.RawValue()}))
				}
			} else if ptrType == reflect.TypeOf(BlankNode{}) {
				if bn, ok := objValue.(*BlankNode); ok {
					newVal.Elem().Set(reflect.ValueOf(*bn))
				} else {
					newVal.Elem().Set(reflect.ValueOf(BlankNode{ID: objValue.RawValue()}))
				}
			} else {
				// For other structs, recursively unmarshal
				if err := u.Unmarshal(objValue.RawValue(), newVal.Interface()); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported pointer type: %s", ptrType.String())
		}
		
		field.Set(newVal)
		
	case reflect.Slice:
		// Check if the object is a collection (rdf:List) or container (rdf:Bag, rdf:Seq, rdf:Alt)
		if res, ok := triple.Object.(*rdf2go.Resource); ok {
			// Handle RDF Collection (rdf:List) - ordered list
			if err := u.unmarshalCollection(res.URI, field); err != nil {
				return err
			}
			return nil
		} else if bn, ok := triple.Object.(*rdf2go.BlankNode); ok {
			// First, check if this is an RDF list (has rdf:first and rdf:rest properties)
			firstTriples := u.graph.All(
				bn,
				rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
				nil,
			)
			restTriples := u.graph.All(
				bn,
				rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
				nil,
			)
			
			// Debug information for blank node detection
			fmt.Printf("Debug: Checking blank node %s for RDF list properties: first=%d, rest=%d\n", 
				bn.String(), len(firstTriples), len(restTriples))
			
			if len(firstTriples) > 0 && len(restTriples) > 0 {
				// This is an RDF list
				fmt.Printf("Debug: Detected RDF list at blank node %s\n", bn.String())
				if err := u.unmarshalCollection(bn.String(), field); err != nil {
					return err
				}
				return nil
			}
			
			// If not an RDF list, check if it's an RDF container (Bag, Seq, Alt)
			containerTypeTriples := u.graph.All(
				bn,
				rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
				nil,
			)
			
			if len(containerTypeTriples) > 0 {
				if typeRes, ok := containerTypeTriples[0].Object.(*rdf2go.Resource); ok {
					typeURI := typeRes.URI
					if typeURI == "http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag" ||
					   typeURI == "http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq" ||
					   typeURI == "http://www.w3.org/1999/02/22-rdf-syntax-ns#Alt" {
						// Handle RDF Container
						if err := u.unmarshalContainer(bn, field); err != nil {
							return err
						}
						return nil
					}
				}
			}
			
			// Check if this is a nested structure with multiple properties
			nestedTriples := u.graph.All(bn, nil, nil)
			if len(nestedTriples) > 0 {
				// This is a nested structure, try to unmarshal it
				if field.Type().Elem().Kind() == reflect.Struct {
					// Create a new instance of the struct
					newElem := reflect.New(field.Type().Elem())
					
					// Unmarshal the nested structure
					for _, nestedTriple := range nestedTriples {
						if predRes, ok := nestedTriple.Predicate.(*rdf2go.Resource); ok {
							predURI := predRes.URI
							
							// Skip rdf:type predicates
							if predURI == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" {
								continue
							}
							
							// Find the field in the struct that matches this predicate
							structType := field.Type().Elem()
							structVal := newElem.Elem()
							
							for i := 0; i < structType.NumField(); i++ {
								structField := structType.Field(i)
								structFieldVal := structVal.Field(i)
								
								// Skip unexported fields
								if !structFieldVal.CanSet() {
									continue
								}
								
								// Get predicate from tag
								tagPredicate := structField.Tag.Get("rdf")
								if tagPredicate == "" {
									continue
								}
								
								if tagPredicate == predURI {
									// Found a matching field, unmarshal the value
									objValue := FromRDF2GoTerm(nestedTriple.Object)
									
									// Set the field value
									if err := u.setFieldValue(structFieldVal, objValue); err != nil {
										return err
									}
								}
							}
						}
					}
					
					// Create a new slice with one element
					newSlice := reflect.MakeSlice(field.Type(), 0, 1)
					newSlice = reflect.Append(newSlice, newElem.Elem())
					field.Set(newSlice)
					return nil
				}
			}
		}
		
		// Handle regular slices (multiple values for the same predicate)
		sliceType := field.Type().Elem()
		sliceVal := reflect.MakeSlice(field.Type(), 0, len(triples))
		
		for _, t := range triples {
			objVal := FromRDF2GoTerm(t.Object)
			
			// Create and set a new value based on the slice element type
			var elemVal reflect.Value
			
			switch sliceType.Kind() {
			case reflect.String:
				elemVal = reflect.ValueOf(objVal.RawValue())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(objVal.RawValue(), 10, 64)
				if err != nil {
					return fmt.Errorf("cannot convert %s to int: %w", objVal.RawValue(), err)
				}
				elemVal = reflect.ValueOf(i).Convert(sliceType)
			case reflect.Struct:
				if sliceType == reflect.TypeOf(Resource{}) {
					if res, ok := objVal.(*Resource); ok {
						elemVal = reflect.ValueOf(*res)
					} else {
						elemVal = reflect.ValueOf(Resource{URI: objVal.RawValue()})
					}
				} else {
					// For other structs, recursively unmarshal
					newElem := reflect.New(sliceType)
					if err := u.Unmarshal(objVal.RawValue(), newElem.Interface()); err != nil {
						return err
					}
					elemVal = newElem.Elem()
				}
			case reflect.Ptr:
				ptrElemType := sliceType.Elem()
				newElem := reflect.New(ptrElemType)
				
				if ptrElemType == reflect.TypeOf(Resource{}) {
					if res, ok := objVal.(*Resource); ok {
						newElem.Elem().Set(reflect.ValueOf(*res))
					} else {
						newElem.Elem().Set(reflect.ValueOf(Resource{URI: objVal.RawValue()}))
					}
				} else {
					// For other types, recursively unmarshal
					if err := u.Unmarshal(objVal.RawValue(), newElem.Interface()); err != nil {
						return err
					}
				}
				
				elemVal = newElem
			default:
				return fmt.Errorf("unsupported slice element type: %s", sliceType.String())
			}
			
			sliceVal = reflect.Append(sliceVal, elemVal)
		}
		
		field.Set(sliceVal)
		
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type().String())
	}
	
	return nil
}

// unmarshalLocalizedField unmarshals a field that contains localized text
func (u *Unmarshaller) unmarshalLocalizedField(subject, predicate string, field reflect.Value) error {
	// Handle different subject types (resource or blank node)
	var subjectTerm rdf2go.Term
	if strings.HasPrefix(subject, "_:") {
		// This is a blank node
		blankNodeID := strings.TrimPrefix(subject, "_:")
		subjectTerm = rdf2go.NewBlankNode(blankNodeID)
	} else {
		// This is a resource
		subjectTerm = rdf2go.NewResource(subject)
	}
	
	triples := u.graph.All(
		subjectTerm,
		rdf2go.NewResource(predicate),
		nil,
	)
	
	if len(triples) == 0 {
		return nil // No value to set
	}
	
	fieldType := field.Type()
	
	// Handle LocalizedString (single language)
	if fieldType == reflect.TypeOf(LocalizedString{}) || fieldType == reflect.TypeOf(&LocalizedString{}) {
		// Use the first triple
		triple := triples[0]
		lit, ok := triple.Object.(*rdf2go.Literal)
		if !ok {
			return fmt.Errorf("expected literal for localized string, got %T", triple.Object)
		}
		
		ls := NewLocalizedString(lit.Value, lit.Language)
		
		if fieldType.Kind() == reflect.Ptr {
			field.Set(reflect.ValueOf(ls))
		} else {
			field.Set(reflect.ValueOf(*ls))
		}
		
		return nil
	}
	
	// Handle LocalizedText (multiple languages)
	if fieldType == reflect.TypeOf(LocalizedText{}) || fieldType == reflect.TypeOf(&LocalizedText{}) {
		lt := NewLocalizedText()
		
		for _, triple := range triples {
			lit, ok := triple.Object.(*rdf2go.Literal)
			if !ok {
				continue // Skip non-literals
			}
			
			lt.Add(lit.Language, lit.Value)
		}
		
		if fieldType.Kind() == reflect.Ptr {
			field.Set(reflect.ValueOf(lt))
		} else {
			field.Set(reflect.ValueOf(*lt))
		}
		
		return nil
	}
	
	return fmt.Errorf("unsupported localized field type: %s", fieldType.String())
}

// implementsRDFUnmarshaler checks if a value implements the RDFUnmarshaler interface
func (u *Unmarshaller) implementsRDFUnmarshaler(v reflect.Value) bool {
	// Get the type of RDFUnmarshaler
	unmarshalerType := reflect.TypeOf((*RDFUnmarshaler)(nil)).Elem()
	
	// Check if the value's type implements RDFUnmarshaler
	return v.Type().Implements(unmarshalerType)
}

// GetResource retrieves a resource by its URI
func (u *Unmarshaller) GetResource(uri string, v interface{}) error {
	return u.Unmarshal(uri, v)
}

// GetResources retrieves all resources of a given type
func (u *Unmarshaller) GetResources(typeURI string, v interface{}) error {
	// v must be a pointer to a slice
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("target must be a non-nil pointer to a slice")
	}
	
	sliceVal := val.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("target must be a pointer to a slice")
	}
	
	// Get the type of the slice elements
	elemType := sliceVal.Type().Elem()
	
	// Find all subjects with the given type
	triples := u.graph.All(nil, rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), rdf2go.NewResource(typeURI))
	
	// Create a new slice with the appropriate capacity
	newSlice := reflect.MakeSlice(sliceVal.Type(), 0, len(triples))
	
	// Unmarshal each resource
	for _, triple := range triples {
		subject := triple.Subject.(*rdf2go.Resource).URI
		
		// Create a new element
		newElem := reflect.New(elemType)
		
		// Unmarshal the resource
		if err := u.Unmarshal(subject, newElem.Interface()); err != nil {
			return err
		}
		
		// Append to the slice
		newSlice = reflect.Append(newSlice, newElem.Elem())
	}
	
	// Set the slice value
	sliceVal.Set(newSlice)
	
	return nil
}

// GetValues retrieves all values for a given subject and predicate
func (u *Unmarshaller) GetValues(subject, predicate string) []Value {
	triples := u.graph.All(
		rdf2go.NewResource(subject),
		rdf2go.NewResource(predicate),
		nil,
	)
	
	values := make([]Value, 0, len(triples))
	for _, triple := range triples {
		values = append(values, FromRDF2GoTerm(triple.Object))
	}
	
	return values
}

// GetValue retrieves a single value for a given subject and predicate
func (u *Unmarshaller) GetValue(subject, predicate string) (Value, error) {
	values := u.GetValues(subject, predicate)
	if len(values) == 0 {
		return nil, fmt.Errorf("no value found for %s %s", subject, predicate)
	}
	return values[0], nil
}

// GetString retrieves a string value for a given subject and predicate
func (u *Unmarshaller) GetString(subject, predicate string) (string, error) {
	value, err := u.GetValue(subject, predicate)
	if err != nil {
		return "", err
	}
	return value.RawValue(), nil
}

// GetInt retrieves an integer value for a given subject and predicate
func (u *Unmarshaller) GetInt(subject, predicate string) (int, error) {
	value, err := u.GetValue(subject, predicate)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value.RawValue())
}

// GetBool retrieves a boolean value for a given subject and predicate
func (u *Unmarshaller) GetBool(subject, predicate string) (bool, error) {
	value, err := u.GetValue(subject, predicate)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(value.RawValue())
}

// GetFloat retrieves a float value for a given subject and predicate
func (u *Unmarshaller) GetFloat(subject, predicate string) (float64, error) {
	value, err := u.GetValue(subject, predicate)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(value.RawValue(), 64)
}

// GetLocalizedText retrieves all language-tagged values for a given subject and predicate
func (u *Unmarshaller) GetLocalizedText(subject, predicate string) *LocalizedText {
	triples := u.graph.All(
		rdf2go.NewResource(subject),
		rdf2go.NewResource(predicate),
		nil,
	)
	
	lt := NewLocalizedText()
	
	for _, triple := range triples {
		lit, ok := triple.Object.(*rdf2go.Literal)
		if !ok {
			continue // Skip non-literals
		}
		
		lt.Add(lit.Language, lit.Value)
	}
	
	return lt
}
