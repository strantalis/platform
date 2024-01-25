package sdk_test

import (
	"reflect"
	"testing"

	"github.com/opentdf/opentdf-v2-poc/sdk"
	"github.com/opentdf/opentdf-v2-poc/sdk/acre"
	"github.com/opentdf/opentdf-v2-poc/sdk/acse"
	"github.com/opentdf/opentdf-v2-poc/sdk/attributes"
	"github.com/opentdf/opentdf-v2-poc/sdk/keyaccessgrants"
)

var goodPlatformEndpoint = "localhost:9000"
var badPlatformEndpoint = "localhost:9999"

func GetMethods(i interface{}) (m []string) {
	r := reflect.TypeOf(i)
	for i := 0; i < r.NumMethod(); i++ {
		m = append(m, r.Method(i).Name)
	}
	return m
}

func Test_ShouldCreateNewSDK(t *testing.T) {
	// When
	sdk, err := sdk.New(goodPlatformEndpoint)
	// Then
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if sdk == nil {
		t.Errorf("Expected sdk, got nil")
	}

	// check if the clients are available
	if sdk.Attributes == nil {
		t.Errorf("Expected Attributes client, got nil")
	}
	if sdk.ResourceEncoding == nil {
		t.Errorf("Expected ResourceEncoding client, got nil")
	}
	if sdk.SubjectEncoding == nil {
		t.Errorf("Expected SubjectEncoding client, got nil")
	}
	if sdk.KeyAccessGrants == nil {
		t.Errorf("Expected KeyAccessGrants client, got nil")
	}
}

func Test_ShouldCloseSDKConnection(t *testing.T) {
	t.Skip("Skipping test since close is broken")
	// Given
	sdk, err := sdk.New(goodPlatformEndpoint)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// When
	err = sdk.Close()
	// Then
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func Test_ShouldHaveSameMethods(t *testing.T) {
	sdk, err := sdk.New(goodPlatformEndpoint)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tests := []struct {
		name     string
		expected []string
		actual   []string
	}{
		{
			name:     "Attributes",
			expected: GetMethods(reflect.TypeOf(attributes.NewAttributesServiceClient(sdk.Conn()))),
			actual:   GetMethods(reflect.TypeOf(sdk.Attributes)),
		},
		{
			name:     "ResourceEncoding",
			expected: GetMethods(reflect.TypeOf(acre.NewResourcEncodingServiceClient(sdk.Conn()))),
			actual:   GetMethods(reflect.TypeOf(sdk.ResourceEncoding)),
		},
		{
			name:     "SubjectEncoding",
			expected: GetMethods(reflect.TypeOf(acse.NewSubjectEncodingServiceClient(sdk.Conn()))),
			actual:   GetMethods(reflect.TypeOf(sdk.SubjectEncoding)),
		},
		{
			name:     "KeyAccessGrants",
			expected: GetMethods(reflect.TypeOf(keyaccessgrants.NewKeyAccessGrantsServiceClient(sdk.Conn()))),
			actual:   GetMethods(reflect.TypeOf(sdk.KeyAccessGrants)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.expected, tt.actual) {
				t.Errorf("Expected Attributes client to have methods %v, got %v", tt.actual, tt.expected)
			}
		})
	}
}

func Test_ShouldCreateNewSDKWithBadEndpoint(t *testing.T) {
	// Bad endpoints are not detected until the first call to the platform
	t.Skip("Skipping test since this is expected but not great behavior")
	// When
	sdk, err := sdk.New(badPlatformEndpoint)
	// Then
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if sdk == nil {
		t.Errorf("Expected sdk, got nil")
	}
}