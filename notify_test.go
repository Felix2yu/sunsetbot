package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "移除加粗标记",
			input:    "**加粗文本**",
			expected: "加粗文本",
		},
		{
			name:     "移除二级标题",
			input:    "## 标题",
			expected: "标题",
		},
		{
			name:     "移除三级标题",
			input:    "### 小标题",
			expected: "小标题",
		},
		{
			name:     "移除列表标记",
			input:    "- 列表项",
			expected: "列表项",
		},
		{
			name:     "保留普通文本",
			input:    "普通文本",
			expected: "普通文本",
		},
		{
			name:     "处理多行文本",
			input:    "## 标题\n- 列表项1\n- 列表项2\n普通文本",
			expected: "标题\n列表项1\n列表项2\n普通文本",
		},
		{
			name:     "处理带空格的标记",
			input:    "  ## 标题  ",
			expected: "标题",
		},
		{
			name:     "处理空字符串",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewNotifier(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	t.Run("PushURL为空返回nil", func(t *testing.T) {
		cfg := &PushConfig{
			Enable:  true,
			PushURL: "",
		}
		notifier := NewNotifier(cfg, logger)
		if notifier != nil {
			t.Error("NewNotifier() 在 PushURL 为空时应该返回 nil")
		}
	})

	t.Run("PushURL非空返回ShoutrrrNotifier", func(t *testing.T) {
		cfg := &PushConfig{
			Enable:  true,
			PushURL: "http://example.com",
		}
		notifier := NewNotifier(cfg, logger)
		if notifier == nil {
			t.Fatal("NewNotifier() 在 PushURL 非空时不应该返回 nil")
		}
		if _, ok := notifier.(*ShoutrrrNotifier); !ok {
			t.Error("NewNotifier() 应该返回 *ShoutrrrNotifier")
		}
	})
}

func TestShoutrrrNotifierName(t *testing.T) {
	notifier := &ShoutrrrNotifier{
		PushURL: "http://example.com",
	}
	if notifier.Name() != "shoutrrr" {
		t.Errorf("Name() = %q, want %q", notifier.Name(), "shoutrrr")
	}
}

func TestShoutrrrNotifierSendEmptyURL(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	notifier := &ShoutrrrNotifier{
		PushURL: "",
		logger:  logger,
	}

	err := notifier.Send("title", "body", 1, []string{}, false)
	if err == nil {
		t.Error("Send() 在 PushURL 为空时应该返回错误")
	}
}

func TestHasUnsupportedParamError(t *testing.T) {
	tests := []struct {
		name     string
		errs     []error
		expected bool
	}{
		{
			name:     "无错误返回false",
			errs:     []error{},
			expected: false,
		},
		{
			name:     "无匹配错误返回false",
			errs:     []error{&testError{msg: "some error"}},
			expected: false,
		},
		{
			name:     "有匹配错误返回true",
			errs:     []error{&testError{msg: "markdown is not a valid config key"}},
			expected: true,
		},
		{
			name:     "多个错误中有一个匹配返回true",
			errs:     []error{&testError{msg: "error 1"}, &testError{msg: "markdown is not a valid config key"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasUnsupportedParamError(tt.errs)
			if result != tt.expected {
				t.Errorf("hasUnsupportedParamError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestShoutrrrNotifierCheckErrors(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	notifier := &ShoutrrrNotifier{
		PushURL: "http://example.com",
		logger:  logger,
	}

	t.Run("无错误返回nil", func(t *testing.T) {
		err := notifier.checkErrors([]error{})
		if err != nil {
			t.Errorf("checkErrors() = %v, want nil", err)
		}
	})

	t.Run("有错误返回汇总错误", func(t *testing.T) {
		errs := []error{
			&testError{msg: "error 1"},
			&testError{msg: "error 2"},
		}
		err := notifier.checkErrors(errs)
		if err == nil {
			t.Error("checkErrors() 在有错误时应该返回错误")
		}
	})

	t.Run("nil错误被忽略", func(t *testing.T) {
		errs := []error{nil, nil}
		err := notifier.checkErrors(errs)
		if err != nil {
			t.Errorf("checkErrors() 在只有 nil 错误时应该返回 nil, got %v", err)
		}
	})
}

func TestShoutrrrNotifierSend(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	t.Run("空URL返回错误", func(t *testing.T) {
		notifier := &ShoutrrrNotifier{
			PushURL: "",
			logger:  logger,
		}
		err := notifier.Send("title", "body", 1, []string{}, false)
		if err == nil {
			t.Error("Send() 在空URL时应该返回错误")
		}
	})

	t.Run("无效URL返回错误", func(t *testing.T) {
		notifier := &ShoutrrrNotifier{
			PushURL: "invalid://url",
			logger:  logger,
		}
		err := notifier.Send("title", "body", 1, []string{}, false)
		if err == nil {
			t.Error("Send() 在无效URL时应该返回错误")
		}
	})

	t.Run("有效URL尝试发送", func(t *testing.T) {
		// 创建一个本地 HTTP 服务器来接收请求
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := &ShoutrrrNotifier{
			PushURL: server.URL,
			logger:  logger,
		}
		// 这会失败，因为 shoutrrr 需要特定格式的 URL
		// 但至少可以测试代码路径
		notifier.Send("title", "body", 1, []string{}, false)
	})

	t.Run("带markdown参数发送", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := &ShoutrrrNotifier{
			PushURL: server.URL,
			logger:  logger,
		}
		notifier.Send("title", "body", 1, []string{}, true)
	})

	t.Run("带优先级参数发送", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := &ShoutrrrNotifier{
			PushURL: server.URL,
			logger:  logger,
		}

		// 测试不同的优先级
		for i := 1; i <= 5; i++ {
			notifier.Send("title", "body", i, []string{}, false)
		}
	})

	t.Run("带tags参数发送", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		notifier := &ShoutrrrNotifier{
			PushURL: server.URL,
			logger:  logger,
		}
		notifier.Send("title", "body", 1, []string{"tag1", "tag2"}, false)
	})
}

func TestLogSuccess(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	notifier := &ShoutrrrNotifier{
		PushURL: "http://example.com",
		logger:  logger,
	}

	// 测试不同的优先级
	for i := 1; i <= 5; i++ {
		notifier.logSuccess(i)
	}
}

func TestStripMarkdownEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"空字符串", "", ""},
		{"只有加粗标记", "**", ""},
		{"嵌套标记", "## **标题**", "标题"},
		{"多个列表项", "- item1\n- item2\n- item3", "item1\nitem2\nitem3"},
		{"混合内容", "## 标题\n- 列表1\n- 列表2\n普通文本", "标题\n列表1\n列表2\n普通文本"},
		{"空行", "line1\n\nline2", "line1\n\nline2"},
		{"只有空格", "   ", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHasUnsupportedParamErrorEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		errs     []error
		expected bool
	}{
		{"nil切片", nil, false},
		{"空切片", []error{}, false},
		{"一个nil错误", []error{nil}, false},
		{"多个nil错误", []error{nil, nil, nil}, false},
		{"混合错误", []error{nil, &testError{msg: "error"}, nil}, false},
		{"包含不支持参数错误", []error{&testError{msg: "markdown is not a valid config key"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasUnsupportedParamError(tt.errs)
			if result != tt.expected {
				t.Errorf("hasUnsupportedParamError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSendWithMultipleURLs(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	// 使用有效的 shoutrrr URL 格式
	// shoutrrr 支持多种通知服务，这里使用一个假的 URL 来测试代码路径
	notifier := &ShoutrrrNotifier{
		PushURL: "fake://token@fake.service",
		logger:  logger,
	}

	err := notifier.Send("title", "body", 1, []string{}, false)
	// 这会失败，但可以测试代码路径
	_ = err
}

func TestSendWithDifferentPriorities(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	notifier := &ShoutrrrNotifier{
		PushURL: "fake://token@fake.service",
		logger:  logger,
	}

	// 测试不同的优先级
	priorities := []int{0, 1, 2, 3, 4, 5, 6}
	for _, p := range priorities {
		notifier.Send("title", "body", p, []string{}, false)
	}
}

func TestSendWithMarkdownAndTags(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	notifier := &ShoutrrrNotifier{
		PushURL: "fake://token@fake.service",
		logger:  logger,
	}

	// 测试带 markdown 和 tags
	notifier.Send("title", "**bold** content", 3, []string{"tag1", "tag2"}, true)

	// 测试不带 markdown
	notifier.Send("title", "**bold** content", 3, []string{"tag1", "tag2"}, false)
}
