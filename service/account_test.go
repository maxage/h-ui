package service

import (
	"testing"
)

func TestValidateNodeAccess(t *testing.T) {
	tests := []struct {
		name        string
		nodeAccess  int64
		expectError bool
	}{
		{
			name:        "Valid single node access",
			nodeAccess:  1,
			expectError: false,
		},
		{
			name:        "Invalid node access value",
			nodeAccess:  3,
			expectError: true,
		},
		{
			name:        "Invalid zero value",
			nodeAccess:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodeAccess(tt.nodeAccess)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetUserNodeAccess(t *testing.T) {
	// 这里需要mock数据库，暂时跳过实际测试
	// 在实际项目中应该使用testify/mock或类似工具
	t.Skip("Requires database mocking")
}