package rdft

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/deiu/rdf2go"
	"github.com/utrack/caisson-go/errors"
)

// unmarshalCollection unmarshals an RDF collection (rdf:List) into a slice
func (u *Unmarshaller) unmarshalCollection(listSubject rdf2go.Term, field reflect.Value) error {
	// Create a new slice with the appropriate type
	sliceType := field.Type()
	newSlice := reflect.MakeSlice(sliceType, 0, 8) // Start with a small capacity

	// Process the list recursively
	if err := u.processRDFList(listSubject, sliceType.Elem(), &newSlice); err != nil {
		return err
	}

	// Set the field value
	field.Set(newSlice)
	return nil
}

// processRDFList recursively processes an RDF list and appends items to the slice
func (u *Unmarshaller) processRDFList(listSubject rdf2go.Term, elemType reflect.Type, slice *reflect.Value) error {
	// Check if this is the end of the list (rdf:nil)
	if listSubject.RawValue() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil" {
		return nil
	}

	// Get the first item in the list
	firstTriples := u.graph.All(
		listSubject,
		rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
		nil,
	)

	if len(firstTriples) == 0 {
		return fmt.Errorf("invalid RDF list: no rdf:first for %s", listSubject.RawValue())
	}

	// Create a new element for the slice
	newElem := reflect.New(elemType).Elem()

	// Handle different types of first items
	switch firstObj := firstTriples[0].Object.(type) {
	case *rdf2go.Literal:
		// For literals, just set the value directly
		if err := u.setFieldValue(newElem, firstObj); err != nil {
			return errors.Wrapf(err, "failed to set field value for literal %s", firstObj.Value)
		}

	case *rdf2go.Resource:
		// For resources, check if it's a struct or a simple value

		// Special handling for Resource type
		if elemType == reflect.TypeOf(Resource{}) {
			// Create a Resource object and set its URI
			resource := Resource{URI: firstObj.URI}
			newElem.Set(reflect.ValueOf(resource))
		} else if elemType.Kind() == reflect.Struct {
			// Unmarshal the resource into the struct
			newPtr := reflect.New(elemType)
			if err := u.unmarshal(firstObj, newPtr.Interface()); err != nil {
				return errors.Wrapf(err, "failed to unmarshal resource %s", firstObj.String())
			}
			newElem.Set(newPtr.Elem())
		} else {
			// For non-struct types, just set the value
			if err := u.setFieldValue(newElem, firstObj); err != nil {
				return errors.Wrapf(err, "failed to set field value for resource %s", firstObj.String())
			}
		}

	case *rdf2go.BlankNode:
		// For blank nodes, check if it has properties (nested structure)
		blankNodeTriples := u.graph.All(firstObj, nil, nil)

		if len(blankNodeTriples) > 0 {
			if elemType.Kind() == reflect.Struct {
				// This is a nested structure, unmarshal the blank node into the struct
				newPtr := reflect.New(elemType)

				// Use the blank node ID as the subject for unmarshalling
				if err := u.unmarshal(firstObj, newPtr.Interface()); err != nil {
					for _, t := range blankNodeTriples {
						fmt.Printf("  %s %s %s\n", t.Subject, t.Predicate, t.Object)
					}

					return errors.Wrapf(err, "failed to unmarshal blank node %s", firstObj.String())
				}
				newElem.Set(newPtr.Elem())
			} else if elemType.Kind() == reflect.Slice {
				// Check if this is a nested RDF list
				firstNestedTriples := u.graph.All(
					firstObj,
					rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
					nil,
				)
				restNestedTriples := u.graph.All(
					firstObj,
					rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
					nil,
				)

				if len(firstNestedTriples) > 0 && len(restNestedTriples) > 0 {
					// This is a nested RDF list

					// Create a new slice for the nested list
					nestedSlice := reflect.MakeSlice(elemType, 0, 8)

					// Process the nested list
					if err := u.processRDFList(firstObj, elemType.Elem(), &nestedSlice); err != nil {
						return errors.Wrapf(err, "failed to process nested RDF list %s", firstObj.String())
					}

					// Set the nested slice as the value of the element
					newElem.Set(nestedSlice)
				} else {
					// Not a nested list, try to unmarshal as a regular value
					firstValue := firstObj
					if err := u.setFieldValue(newElem, firstValue); err != nil {
						return errors.Wrapf(err, "failed to set field value for first item %s", firstObj.String())
					}
				}
			} else {
				// For non-struct types, just set the value
				if err := u.setFieldValue(newElem, firstObj); err != nil {
					return errors.Wrapf(err, "failed to set field value for non-struct type %s", firstObj.String())
				}
			}
		} else {
			// For empty blank nodes, just set the value
			if err := u.setFieldValue(newElem, firstObj); err != nil {
				return errors.Wrapf(err, "failed to set field value for empty blank node %s", firstObj.String())
			}
		}

	default:
		// For other types, just set the value
		if err := u.setFieldValue(newElem, firstTriples[0].Object); err != nil {
			return errors.Wrapf(err, "failed to set field value for other type %s", firstTriples[0].Object.String())
		}
	}

	// Append the element to the slice
	*slice = reflect.Append(*slice, newElem)

	// Get the rest of the list
	restTriples := u.graph.All(
		listSubject,
		rdf2go.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
		nil,
	)

	if len(restTriples) == 0 {
		return fmt.Errorf("invalid RDF list: no rdf:rest for %s", listSubject.String())
	}

	// Process the rest of the list
	rest := restTriples[0].Object
	if rest.RawValue() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#nil" {
		// End of the list
		return nil
	}
	return u.processRDFList(rest, elemType, slice)
}

// unmarshalContainer unmarshals an RDF container (rdf:Bag, rdf:Seq, rdf:Alt) into a slice
func (u *Unmarshaller) unmarshalContainer(containerNode rdf2go.Term, field reflect.Value) error {
	// Get all triples with the container as subject
	containerTriples := u.graph.All(containerNode, nil, nil)

	// Filter out the type triple
	var memberTriples []*rdf2go.Triple
	for _, triple := range containerTriples {
		predURI := triple.Predicate.(*rdf2go.Resource).URI
		if !strings.HasPrefix(predURI, "http://www.w3.org/1999/02/22-rdf-syntax-ns#_") {
			continue
		}
		memberTriples = append(memberTriples, triple)
	}

	// For rdf:Seq, sort by predicate (_1, _2, _3, etc.)
	sort.Slice(memberTriples, func(i, j int) bool {
		// Extract the index from the predicate URI
		predI := memberTriples[i].Predicate.(*rdf2go.Resource).URI
		predJ := memberTriples[j].Predicate.(*rdf2go.Resource).URI

		// Extract the numeric part after rdf:_
		idxI := strings.TrimPrefix(predI, "http://www.w3.org/1999/02/22-rdf-syntax-ns#_")
		idxJ := strings.TrimPrefix(predJ, "http://www.w3.org/1999/02/22-rdf-syntax-ns#_")

		// Convert to integers and compare
		numI, _ := strconv.Atoi(idxI)
		numJ, _ := strconv.Atoi(idxJ)
		return numI < numJ
	})

	// Create a new slice with the appropriate capacity
	sliceType := field.Type()
	newSlice := reflect.MakeSlice(sliceType, 0, len(memberTriples))

	// Add each member to the slice
	for _, triple := range memberTriples {
		// Create a new element for the slice
		elemType := sliceType.Elem()
		newElem := reflect.New(elemType).Elem()

		// Get the value
		objValue := triple.Object

		// Set the value based on the element type
		if err := u.setFieldValue(newElem, objValue); err != nil {
			return err
		}

		// Append the element to the slice
		newSlice = reflect.Append(newSlice, newElem)
	}

	// Set the field value
	field.Set(newSlice)
	return nil
}

// setFieldValue sets a field value based on its type and the RDF value
func (u *Unmarshaller) setFieldValue(field reflect.Value, object rdf2go.Term) error {
	switch field.Kind() {
	case reflect.String:
		// Check if the object is a resource - if so, this is a type mismatch
		if _, isResource := object.(*Resource); isResource {
			return fmt.Errorf("type mismatch: cannot unmarshal Resource into string field, use rdft.Resource instead")
		}
		field.SetString(object.RawValue())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(object.RawValue(), 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to int: %w", object.RawValue(), err)
		}
		field.SetInt(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(object.RawValue(), 64)
		if err != nil {
			return fmt.Errorf("cannot convert %s to float: %w", object.RawValue(), err)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(object.RawValue())
		if err != nil {
			return fmt.Errorf("cannot convert %s to bool: %w", object.RawValue(), err)
		}
		field.SetBool(b)
	case reflect.Struct:
		// For Resource type
		if field.Type() == reflect.TypeOf(Resource{}) {
			if res, ok := object.(*Resource); ok {
				field.Set(reflect.ValueOf(*res))
			} else {
				return fmt.Errorf("cannot convert %v to Resource", object)
			}
		} else {
			// For other structs, recursively unmarshal
			newVal := reflect.New(field.Type())
			if err := u.Unmarshal(object.RawValue(), newVal.Interface()); err != nil {
				return err
			}
			field.Set(newVal.Elem())
		}
	case reflect.Ptr:
		// Create a new instance of the pointed-to type
		ptrType := field.Type().Elem()
		newVal := reflect.New(ptrType)

		// Set the value based on the pointed-to type
		if err := u.setFieldValue(newVal.Elem(), object); err != nil {
			return err
		}

		field.Set(newVal)
	case reflect.Slice:
		// If the value is a resource or blank node, it might be a nested collection or container
		if res, ok := object.(*Resource); ok {
			return u.unmarshalCollection(res, field)
		} else if bn, ok := object.(rdf2go.BlankNode); ok {
			// Check if this is a container
			containerTypeTriples := u.graph.All(
				rdf2go.NewBlankNode(bn.ID),
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
						return u.unmarshalContainer(rdf2go.NewBlankNode(bn.ID), field)
					}
				}
			}
		}

		// If not a collection or container, create a single-element slice
		sliceType := field.Type()
		newSlice := reflect.MakeSlice(sliceType, 0, 1)

		// Create a new element
		elemType := sliceType.Elem()
		newElem := reflect.New(elemType).Elem()

		// Set the element value
		if err := u.setFieldValue(newElem, object); err != nil {
			return err
		}

		// Append to the slice
		newSlice = reflect.Append(newSlice, newElem)
		field.Set(newSlice)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type().String())
	}

	return nil
}
