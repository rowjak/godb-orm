package generator

import (
	"testing"
)

func TestNamingConverter_ToGoFieldName(t *testing.T) {
	nc := NewNamingConverter()

	tests := []struct {
		input    string
		expected string
	}{
		{"id", "ID"},
		{"user_id", "UserID"},
		{"created_at", "CreatedAt"},
		{"first_name", "FirstName"},
		{"api_key", "APIKey"},
		{"json_data", "JSONData"},
		{"http_status", "HTTPStatus"},
		{"url", "URL"},
		{"some_column", "SomeColumn"},
		{"email", "Email"},
		{"uuid", "UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := nc.ToGoFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("ToGoFieldName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNamingConverter_ToGoStructName(t *testing.T) {
	nc := NewNamingConverter()

	tests := []struct {
		input    string
		expected string
	}{
		{"users", "User"},
		{"order_items", "OrderItem"},
		{"categories", "Category"},
		{"posts", "Post"},
		{"user_profiles", "UserProfile"},
		{"addresses", "Address"},
		{"people", "Person"},
		{"children", "Child"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := nc.ToGoStructName(tt.input)
			if result != tt.expected {
				t.Errorf("ToGoStructName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNamingConverter_ToFileName(t *testing.T) {
	nc := NewNamingConverter()

	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users.go"},
		{"UserProfile", "user_profile.go"},
		{"order_items", "order_items.go"},
		{"APIKey", "api_key.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := nc.ToFileName(tt.input)
			if result != tt.expected {
				t.Errorf("ToFileName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "user"},
		{"posts", "post"},
		{"categories", "category"},
		{"boxes", "box"},
		{"classes", "class"},
		{"people", "person"},
		{"children", "child"},
		{"addresses", "address"},
		{"statuses", "status"},
		{"leaves", "leaf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := singularize(tt.input)
			if result != tt.expected {
				t.Errorf("singularize(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHandleAcronyms(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserId", "UserID"},
		{"ApiUrl", "APIURL"},
		{"HttpStatus", "HTTPStatus"},
		{"JsonData", "JSONData"},
		{"Uuid", "UUID"},
		{"IpAddress", "IPAddress"},
		{"DbConnection", "DBConnection"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := handleAcronyms(tt.input)
			if result != tt.expected {
				t.Errorf("handleAcronyms(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
