package nameutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxLengthAndPaddingSpaces(t *testing.T) {
	assert.Equal(t, 11, MaxLength([]string{"short", "hello world", "mid"}))
	assert.Equal(t, 0, MaxLength(nil))
	assert.Equal(t, "   ", PaddingSpaces(3))
}

func TestCaseConversions(t *testing.T) {
	assert.Equal(t, "UserService", ToCamel("user_service"))
	assert.Equal(t, "HttpServer2", ToCamel("http-server_2"))
	assert.Equal(t, "userService", ToLowerCamel("user_service"))
	assert.Equal(t, "userService2", ToLowerCamel("UserService2"))
	assert.Equal(t, "user_service", ToSnake("UserService"))
	assert.Equal(t, "USER_SERVICE", ToScreamingSnake("userService"))
	assert.Equal(t, "T_ITEM", ToScreamingSnake("TItem"))
	assert.Equal(t, "", ToCamel(""))
	assert.Equal(t, "", ToLowerCamel(""))
	assert.Equal(t, "", ToSnake(""))
	assert.Equal(t, "", ToScreamingSnake(""))
}

func TestCaseMatchers(t *testing.T) {
	assert.True(t, IsSnakeCase("user_name_2"))
	assert.False(t, IsSnakeCase("UserName"))
	assert.False(t, IsSnakeCase("user__name"))
	assert.False(t, IsSnakeCase("user_name_"))

	assert.True(t, IsScreamingSnakeCase("USER_NAME_2"))
	assert.False(t, IsScreamingSnakeCase("USER_Name"))

	assert.True(t, IsCamelCase("UserName"))
	assert.True(t, IsCamelCase("TItem"))
	assert.False(t, IsCamelCase("userName"))
	assert.False(t, IsCamelCase("User_Name"))

	assert.True(t, IsLowerCamelCase("userName"))
	assert.True(t, IsLowerCamelCase("userName2"))
	assert.False(t, IsLowerCamelCase("UserName"))
	assert.False(t, IsLowerCamelCase("user_name"))
}

func TestSplitWordsMixedInput(t *testing.T) {
	assert.Equal(t, []string{"HTTP", "Server", "2", "Test"}, splitWords("HTTPServer2-Test"))
	assert.True(t, strings.Contains(ToScreamingSnake("HTTPServer2-Test"), "HTTP_SERVER_2_TEST"))
}

func TestCutKindSuffix(t *testing.T) {
	for _, test := range []struct {
		name, suffix, body, version string
		found                       bool
	}{
		{"OrderService", "Service", "Order", "", true},
		{"OrderServiceV1", "Service", "Order", "V1", true},
		{"OrderServiceV10", "Service", "Order", "V10", true},
		{"OrderServiceV999999999999999999999", "Service", "Order", "V999999999999999999999", true},
		{"OrderApiServiceV2", "ApiService", "Order", "V2", true},
		{"OrderV2Service", "Service", "OrderV2", "", true},
		{"OrderV2EventV10", "Event", "OrderV2", "V10", true},
		{"ServiceV2", "Service", "", "V2", true},
		{"OrderServiceV0", "Service", "OrderServiceV0", "", false},
		{"OrderServiceV01", "Service", "OrderServiceV01", "", false},
		{"OrderServiceV", "Service", "OrderServiceV", "", false},
		{"OrderServicev2", "Service", "OrderServicev2", "", false},
		{"OrderServiceV2Beta", "Service", "OrderServiceV2Beta", "", false},
		{"OrderServiceV٢", "Service", "OrderServiceV٢", "", false},
		{"OrderServiceV2", "ApiService", "OrderServiceV2", "", false},
	} {
		t.Run(test.name+"/"+test.suffix, func(t *testing.T) {
			body, version, found := CutKindSuffix(test.name, test.suffix)
			assert.Equal(t, test.body, body)
			assert.Equal(t, test.version, version)
			assert.Equal(t, test.found, found)
		})
	}
}
