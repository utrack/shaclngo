package rdft_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/deiu/rdf2go"
	"github.com/utrack/shaclngo/rdft"
)

// TestCollections tests unmarshalling RDF collections (ordered lists) into Go slices
func TestCollections(t *testing.T) {
	// Load the test graph
	g, err := loadCollectionsGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define a struct for a person with interests (RDF Collection)
	type Person struct {
		URI       rdft.Resource `rdf:"@id"`
		Type      rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label     string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Age       int           `rdf:"http://example.org/age"`
		Interests []string      `rdf:"http://example.org/interests"`
	}

	// Unmarshal Person1 with a list of interests
	var person1 Person
	err = u.Unmarshal("http://example.org/Person1", &person1)
	if err != nil {
		t.Fatalf("Failed to unmarshal Person1: %v", err)
	}

	// Verify the interests were correctly unmarshalled as a slice
	expectedInterests := []string{"Programming", "Reading", "Hiking"}
	if len(person1.Interests) != len(expectedInterests) {
		t.Errorf("Expected %d interests, got %d", len(expectedInterests), len(person1.Interests))
	}

	for i, interest := range expectedInterests {
		if i < len(person1.Interests) && person1.Interests[i] != interest {
			t.Errorf("Expected interest %d to be %s, got %s", i, interest, person1.Interests[i])
		}
	}

	fmt.Printf("Successfully unmarshalled Person1 with interests: %v\n", person1.Interests)
}

// TestContainers tests unmarshalling RDF containers (Bag, Seq, Alt) into Go slices
func TestContainers(t *testing.T) {
	// Load the test graph
	g, err := loadCollectionsGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define structs for people with different container types
	type PersonWithSkills struct {
		URI    rdft.Resource `rdf:"@id"`
		Type   rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label  string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Age    int           `rdf:"http://example.org/age"`
		Skills []string      `rdf:"http://example.org/skills"`
	}

	type PersonWithJobs struct {
		URI          rdft.Resource `rdf:"@id"`
		Type         rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label        string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Age          int           `rdf:"http://example.org/age"`
		PreviousJobs []string      `rdf:"http://example.org/previousJobs"`
	}

	type PersonWithContacts struct {
		URI            rdft.Resource `rdf:"@id"`
		Type           rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label          string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Age            int           `rdf:"http://example.org/age"`
		ContactMethods []string      `rdf:"http://example.org/contactMethods"`
	}

	// Test Bag container (unordered)
	var person2 PersonWithSkills
	err = u.Unmarshal("http://example.org/Person2", &person2)
	if err != nil {
		t.Fatalf("Failed to unmarshal Person2 with skills bag: %v", err)
	}

	expectedSkills := []string{"Python", "Go", "JavaScript"}
	if len(person2.Skills) != len(expectedSkills) {
		t.Errorf("Expected %d skills, got %d", len(expectedSkills), len(person2.Skills))
	}

	// Since Bag is unordered, we just check that all expected items are present
	skillsMap := make(map[string]bool)
	for _, skill := range person2.Skills {
		skillsMap[skill] = true
	}

	for _, skill := range expectedSkills {
		if !skillsMap[skill] {
			t.Errorf("Expected skill %s not found in unmarshalled skills", skill)
		}
	}

	fmt.Printf("Successfully unmarshalled Person2 with skills bag: %v\n", person2.Skills)

	// Test Seq container (ordered)
	var person3 PersonWithJobs
	err = u.Unmarshal("http://example.org/Person3", &person3)
	if err != nil {
		t.Fatalf("Failed to unmarshal Person3 with jobs sequence: %v", err)
	}

	expectedJobs := []string{"Software Engineer", "Senior Developer", "Tech Lead"}
	if len(person3.PreviousJobs) != len(expectedJobs) {
		t.Errorf("Expected %d jobs, got %d", len(expectedJobs), len(person3.PreviousJobs))
	}

	for i, job := range expectedJobs {
		if i < len(person3.PreviousJobs) && person3.PreviousJobs[i] != job {
			t.Errorf("Expected job %d to be %s, got %s", i, job, person3.PreviousJobs[i])
		}
	}

	fmt.Printf("Successfully unmarshalled Person3 with jobs sequence: %v\n", person3.PreviousJobs)

	// Test Alt container (alternatives)
	var person4 PersonWithContacts
	err = u.Unmarshal("http://example.org/Person4", &person4)
	if err != nil {
		t.Fatalf("Failed to unmarshal Person4 with contact alternatives: %v", err)
	}

	expectedContacts := []string{"email: alice@example.com", "phone: +1-555-123-4567", "twitter: @alicew"}
	if len(person4.ContactMethods) != len(expectedContacts) {
		t.Errorf("Expected %d contact methods, got %d", len(expectedContacts), len(person4.ContactMethods))
	}

	// For Alt container, we just check that all items are present
	contactsMap := make(map[string]bool)
	for _, contact := range person4.ContactMethods {
		contactsMap[contact] = true
	}

	for _, contact := range expectedContacts {
		if !contactsMap[contact] {
			t.Errorf("Expected contact method %s not found in unmarshalled contacts", contact)
		}
	}

	fmt.Printf("Successfully unmarshalled Person4 with contact alternatives: %v\n", person4.ContactMethods)
}

// TestNestedCollections tests unmarshalling nested RDF collections and containers
func TestNestedCollections(t *testing.T) {
	// Load the test graph
	g, err := loadCollectionsGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define structs for nested collections
	type TeamMember struct {
		Type   rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label  string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Role   string        `rdf:"http://example.org/role"`
		Skills []string      `rdf:"http://example.org/skills"`
	}

	type Project struct {
		URI         rdft.Resource `rdf:"@id"`
		Type        rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label       string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		TeamMembers []TeamMember  `rdf:"http://example.org/teamMembers"`
	}

	// Debug: Print all triples related to Project1 and follow the entire graph structure
	fmt.Println("DEBUG: Tracing the entire RDF graph structure for Project1:")
	
	// Helper function to recursively trace through blank nodes
	var traceBlankNode func(bn *rdf2go.BlankNode, depth int)
	traceBlankNode = func(bn *rdf2go.BlankNode, depth int) {
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%sBlank node: %s\n", indent, bn.String())
		
		// Get all triples with this blank node as subject
		bnTriples := g.All(bn, nil, nil)
		for _, triple := range bnTriples {
			fmt.Printf("%s  %s %s %s\n", indent, triple.Subject, triple.Predicate, triple.Object)
			
			// If the object is another blank node, follow it recursively
			if objBn, ok := triple.Object.(*rdf2go.BlankNode); ok {
				traceBlankNode(objBn, depth+1)
			}
		}
	}
	
	// Start with Project1
	project1Triples := g.All(rdf2go.NewResource("http://example.org/Project1"), nil, nil)
	for _, triple := range project1Triples {
		fmt.Printf("Root: %s %s %s\n", triple.Subject, triple.Predicate, triple.Object)
		
		// If the object is a blank node, trace it
		if bn, ok := triple.Object.(*rdf2go.BlankNode); ok {
			traceBlankNode(bn, 1)
		}
	}

	// Unmarshal Project1 with nested team members and their skills
	var project1 Project
	err = u.Unmarshal("http://example.org/Project1", &project1)
	if err != nil {
		t.Fatalf("Failed to unmarshal Project1: %v", err)
	}

	// Verify the project was correctly unmarshalled with nested collections
	if len(project1.TeamMembers) != 2 {
		t.Errorf("Expected 2 team members, got %d", len(project1.TeamMembers))
	}

	// Check first team member
	if project1.TeamMembers[0].Label != "Eve Brown" {
		t.Errorf("Expected first team member to be Eve Brown, got %s", project1.TeamMembers[0].Label)
	}

	expectedSkills1 := []string{"HTML", "CSS", "JavaScript"}
	if len(project1.TeamMembers[0].Skills) != len(expectedSkills1) {
		t.Errorf("Expected %d skills for first team member, got %d", len(expectedSkills1), len(project1.TeamMembers[0].Skills))
	}

	// Check second team member
	if project1.TeamMembers[1].Label != "Charlie Davis" {
		t.Errorf("Expected second team member to be Charlie Davis, got %s", project1.TeamMembers[1].Label)
	}

	expectedSkills2 := []string{"Go", "SQL", "Docker"}
	if len(project1.TeamMembers[1].Skills) != len(expectedSkills2) {
		t.Errorf("Expected %d skills for second team member, got %d", len(expectedSkills2), len(project1.TeamMembers[1].Skills))
	}

	fmt.Printf("Successfully unmarshalled Project1 with nested collections:\n")
	fmt.Printf("  Team Member 1: %s, Role: %s, Skills: %v\n", 
		project1.TeamMembers[0].Label, project1.TeamMembers[0].Role, project1.TeamMembers[0].Skills)
	fmt.Printf("  Team Member 2: %s, Role: %s, Skills: %v\n", 
		project1.TeamMembers[1].Label, project1.TeamMembers[1].Role, project1.TeamMembers[1].Skills)
}

// TestPersonCollection tests unmarshalling a collection of Person references
func TestPersonCollection(t *testing.T) {
	// Load the test graph
	g, err := loadCollectionsGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define a struct for a Person
	type Person struct {
		URI       rdft.Resource `rdf:"@id"`
		Type      rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label     string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Age       int           `rdf:"http://example.org/age"`
	}

	// Define a struct for a Team with a collection of Person references
	type Team struct {
		URI     rdft.Resource   `rdf:"@id"`
		Type    rdft.Resource   `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label   string          `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Members []rdft.Resource `rdf:"http://example.org/members"`
	}

	// Unmarshal Team1 with a collection of Person references
	var team1 Team
	err = u.Unmarshal("http://example.org/Team1", &team1)
	if err != nil {
		t.Fatalf("Failed to unmarshal Team1: %v", err)
	}

	// Verify the team was correctly unmarshalled
	if team1.Label != "Development Team" {
		t.Errorf("Expected team label to be Development Team, got %s", team1.Label)
	}

	// Verify the members collection was correctly unmarshalled
	expectedMembers := []string{
		"http://example.org/Person1",
		"http://example.org/Person2",
		"http://example.org/Person3",
		"http://example.org/Person4",
	}

	if len(team1.Members) != len(expectedMembers) {
		t.Errorf("Expected %d team members, got %d", len(expectedMembers), len(team1.Members))
	}

	// Check each member URI
	for i, expectedURI := range expectedMembers {
		if i < len(team1.Members) && team1.Members[i].URI != expectedURI {
			t.Errorf("Expected member %d to be %s, got %s", i, expectedURI, team1.Members[i].URI)
		}
	}

	// Now unmarshal each person referenced in the team
	persons := make([]Person, len(team1.Members))
	personLabels := []string{"John Doe", "Jane Smith", "Bob Johnson", "Alice Williams"}
	personAges := []int{30, 28, 35, 32}

	for i, memberURI := range team1.Members {
		// Create a new Person instance
		var person Person
		
		// Set the URI field explicitly
		person.URI = memberURI
		
		// Unmarshal the person data
		err = u.Unmarshal(memberURI.URI, &person)
		if err != nil {
			t.Fatalf("Failed to unmarshal Person%d: %v", i+1, err)
		}
		
		// Store the person in the slice
		persons[i] = person

		// Verify each person's properties
		if persons[i].Label != personLabels[i] {
			t.Errorf("Expected Person%d label to be %s, got %s", i+1, personLabels[i], persons[i].Label)
		}

		if persons[i].Age != personAges[i] {
			t.Errorf("Expected Person%d age to be %d, got %d", i+1, personAges[i], persons[i].Age)
		}

		// Verify the URI was correctly unmarshalled
		if persons[i].URI.URI != expectedMembers[i] {
			t.Errorf("Expected Person%d URI to be %s, got %s", i+1, expectedMembers[i], persons[i].URI.URI)
		}
	}

	fmt.Printf("Successfully unmarshalled Team1 with person references:\n")
	fmt.Printf("  Team: %s\n", team1.Label)
	fmt.Printf("  Members: %d people\n", len(team1.Members))
	for i, person := range persons {
		fmt.Printf("  Person%d: %s, Age: %d, URI: %s\n", i+1, person.Label, person.Age, person.URI.URI)
	}
}

// TestMixedCollectionsAndContainers tests unmarshalling mixed RDF collections and containers
func TestMixedCollectionsAndContainers(t *testing.T) {
	// Load the test graph
	g, err := loadCollectionsGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}

	// Create an unmarshaller
	u := rdft.NewUnmarshaller(g)

	// Define structs for mixed collections and containers
	type Team struct {
		Type    rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label   string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Members []string      `rdf:"http://example.org/members"`
	}

	type Manager struct {
		Type  rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Teams []Team        `rdf:"http://example.org/teams"`
	}

	type Department struct {
		URI     rdft.Resource `rdf:"@id"`
		Type    rdft.Resource `rdf:"http://www.w3.org/1999/02/22-rdf-syntax-ns#type"`
		Label   string        `rdf:"http://www.w3.org/2000/01/rdf-schema#label"`
		Manager Manager       `rdf:"http://example.org/manager"`
	}

	// Unmarshal Department1 with mixed collections and containers
	var department1 Department
	err = u.Unmarshal("http://example.org/Department1", &department1)
	if err != nil {
		t.Fatalf("Failed to unmarshal Department1: %v", err)
	}

	// Verify the department was correctly unmarshalled with mixed collections and containers
	if department1.Label != "Engineering" {
		t.Errorf("Expected department label to be Engineering, got %s", department1.Label)
	}

	if department1.Manager.Label != "Dave Wilson" {
		t.Errorf("Expected manager to be Dave Wilson, got %s", department1.Manager.Label)
	}

	if len(department1.Manager.Teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(department1.Manager.Teams))
	}

	// Check first team
	if department1.Manager.Teams[0].Label != "Frontend" {
		t.Errorf("Expected first team to be Frontend, got %s", department1.Manager.Teams[0].Label)
	}

	expectedMembers1 := []string{"Alice", "Bob", "Charlie"}
	if len(department1.Manager.Teams[0].Members) != len(expectedMembers1) {
		t.Errorf("Expected %d members in first team, got %d", len(expectedMembers1), len(department1.Manager.Teams[0].Members))
	}

	// Check second team
	if department1.Manager.Teams[1].Label != "Backend" {
		t.Errorf("Expected second team to be Backend, got %s", department1.Manager.Teams[1].Label)
	}

	expectedMembers2 := []string{"Dave", "Eve", "Frank"}
	if len(department1.Manager.Teams[1].Members) != len(expectedMembers2) {
		t.Errorf("Expected %d members in second team, got %d", len(expectedMembers2), len(department1.Manager.Teams[1].Members))
	}

	fmt.Printf("Successfully unmarshalled Department1 with mixed collections and containers:\n")
	fmt.Printf("  Department: %s\n", department1.Label)
	fmt.Printf("  Manager: %s\n", department1.Manager.Label)
	fmt.Printf("  Team 1: %s, Members: %v\n", department1.Manager.Teams[0].Label, department1.Manager.Teams[0].Members)
	fmt.Printf("  Team 2: %s, Members: %v\n", department1.Manager.Teams[1].Label, department1.Manager.Teams[1].Members)
}

// Helper function to load the collections test graph
func loadCollectionsGraph() (*rdf2go.Graph, error) {
	// Use a fixed base URI
	g := rdf2go.NewGraph("http://example.org/", false)

	// Open the file
	f, err := os.Open("testfiles/collections.ttl")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the Turtle file
	err = g.Parse(f, "text/turtle")
	if err != nil {
		return nil, err
	}

	return g, nil
}
