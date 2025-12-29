package service

import (
	"testing"
)

func TestAppendStorageNotes(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		notes    []string
		expected string
	}{
		{
			name:     "空已有错误，空备注",
			existing: "",
			notes:    []string{},
			expected: "",
		},
		{
			name:     "空已有错误，有备注",
			existing: "",
			notes:    []string{"note1", "note2"},
			expected: "note1; note2",
		},
		{
			name:     "有已有错误，空备注",
			existing: "existing error",
			notes:    []string{},
			expected: "existing error",
		},
		{
			name:     "有已有错误，有备注",
			existing: "existing error",
			notes:    []string{"note1"},
			expected: "existing error; note1",
		},
		{
			name:     "空白已有错误，有备注",
			existing: "   ",
			notes:    []string{"note1"},
			expected: "note1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendStorageNotes(tt.existing, tt.notes)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestComputeInputBaseName(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "空数据",
			data:     []byte{},
			expected: "d41d8cd98f00b204e9800998ecf8427e", // MD5 of empty string
		},
		{
			name:     "Hello",
			data:     []byte("Hello"),
			expected: "8b1a9953c4611296a827abf8c47804d7", // MD5 of "Hello"
		},
		{
			name:     "测试数据",
			data:     []byte("test data"),
			expected: "eb733a00c0c9d336e65691a37ab54293", // MD5 of "test data"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeInputBaseName(tt.data)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildOutputBaseName(t *testing.T) {
	tests := []struct {
		name        string
		modelName   string
		idx         int
		wantPrefix  string
		wantNoEmpty bool
	}{
		{
			name:        "正常模型名",
			modelName:   "test-model",
			idx:         0,
			wantPrefix:  "test-model_",
			wantNoEmpty: true,
		},
		{
			name:        "空模型名",
			modelName:   "",
			idx:         1,
			wantPrefix:  "model_",
			wantNoEmpty: true,
		},
		{
			name:        "超长模型名",
			modelName:   "this-is-a-very-long-model-name-that-exceeds-32-characters",
			idx:         2,
			wantPrefix:  "this-is-a-very-long-model-name-t_",
			wantNoEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOutputBaseName(tt.modelName, tt.idx)

			if tt.wantNoEmpty && result == "" {
				t.Error("expected non-empty result")
			}

			if tt.wantPrefix != "" && len(result) > len(tt.wantPrefix) {
				prefix := result[:len(tt.wantPrefix)]
				if prefix != tt.wantPrefix {
					t.Errorf("expected prefix %q, got %q", tt.wantPrefix, prefix)
				}
			}
		})
	}
}

func TestNewGenerationService(t *testing.T) {
	// 测试创建服务实例
	svc := NewGenerationService(nil, nil)

	if svc == nil {
		t.Fatal("expected service to be created")
	}

	if svc.repo != nil {
		t.Error("expected repo to be nil")
	}

	if svc.storage != nil {
		t.Error("expected storage to be nil")
	}

	if svc.notifyFunc != nil {
		t.Error("expected notifyFunc to be nil")
	}
}

func TestSetNotifyFunc(t *testing.T) {
	svc := NewGenerationService(nil, nil)

	called := false
	testFunc := func(clientID string, recordID uint, status string, errMsg string) {
		called = true
	}

	svc.SetNotifyFunc(testFunc)

	if svc.notifyFunc == nil {
		t.Fatal("expected notifyFunc to be set")
	}

	// 测试调用
	svc.notifyFunc("test-client", 1, "success", "")
	if !called {
		t.Error("expected notifyFunc to be called")
	}
}

func TestNotifyComplete(t *testing.T) {
	t.Run("有通知函数且有 clientID", func(t *testing.T) {
		svc := NewGenerationService(nil, nil)

		notified := false
		receivedClientID := ""
		receivedRecordID := uint(0)
		receivedStatus := ""
		receivedErrMsg := ""

		svc.SetNotifyFunc(func(clientID string, recordID uint, status string, errMsg string) {
			notified = true
			receivedClientID = clientID
			receivedRecordID = recordID
			receivedStatus = status
			receivedErrMsg = errMsg
		})

		svc.notifyComplete("test-client", 123, "success", "test error")

		if !notified {
			t.Error("expected notification")
		}
		if receivedClientID != "test-client" {
			t.Errorf("expected clientID %q, got %q", "test-client", receivedClientID)
		}
		if receivedRecordID != 123 {
			t.Errorf("expected recordID %d, got %d", 123, receivedRecordID)
		}
		if receivedStatus != "success" {
			t.Errorf("expected status %q, got %q", "success", receivedStatus)
		}
		if receivedErrMsg != "test error" {
			t.Errorf("expected errMsg %q, got %q", "test error", receivedErrMsg)
		}
	})

	t.Run("空 clientID 不通知", func(t *testing.T) {
		svc := NewGenerationService(nil, nil)

		notified := false
		svc.SetNotifyFunc(func(clientID string, recordID uint, status string, errMsg string) {
			notified = true
		})

		svc.notifyComplete("", 123, "success", "")

		if notified {
			t.Error("expected no notification for empty clientID")
		}
	})

	t.Run("空白 clientID 不通知", func(t *testing.T) {
		svc := NewGenerationService(nil, nil)

		notified := false
		svc.SetNotifyFunc(func(clientID string, recordID uint, status string, errMsg string) {
			notified = true
		})

		svc.notifyComplete("   ", 123, "success", "")

		if notified {
			t.Error("expected no notification for whitespace clientID")
		}
	})

	t.Run("无通知函数时不崩溃", func(t *testing.T) {
		svc := NewGenerationService(nil, nil)

		// 不应该 panic
		svc.notifyComplete("test-client", 123, "success", "")
	})
}
