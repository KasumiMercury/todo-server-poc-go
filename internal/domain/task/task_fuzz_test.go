package task

import (
	"errors"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/KasumiMercury/todo-server-poc-go/internal/domain/user"
)

// FuzzValidateTitle tests the validateTitle function with various inputs
func FuzzValidateTitle(f *testing.F) {
	testCases := []string{
		"",                                    // Empty string
		"   ",                                 // Whitespace only
		"Valid Title",                         // Normal ASCII
		"日本語タイトル",                             // Japanese characters
		"🚀 Emoji Title 🌟",                     // Emojis
		"Mixed 日本語 and English",               // Mixed languages
		strings.Repeat("a", MaxTitleLength),   // Exactly max length (ASCII)
		strings.Repeat("a", MaxTitleLength+1), // Over max length (ASCII)
		strings.Repeat("あ", MaxTitleLength),   // Exactly max length (Japanese)
		strings.Repeat("あ", MaxTitleLength+1), // Over max length (Japanese)
		strings.Repeat("🚀", MaxTitleLength),   // Exactly max length (Emoji)
		strings.Repeat("🚀", MaxTitleLength+1), // Over max length (Emoji)
		"\x00\x01\x02",                        // Control characters
		"\u200B\u200C\u200D",                  // Zero-width characters
		"Ñoël",                                // Accented characters
		"世界",                                  // Common Japanese words
		"🌍🌎🌏",                                 // Multiple emojis
		"\t\n\r\f\v",                          // Various whitespace
		strings.Repeat("x", 254) + "あ",        // ASCII + 1 Japanese char at boundary
		// Multibyte characters from various languages
		"你好世界",                     // Chinese characters (3 bytes each)
		"안녕하세요 세계",                 // Korean characters (3 bytes each)
		"مرحبا بالعالم",            // Arabic characters (RTL, 2-3 bytes each)
		"Привет мир",               // Russian characters (2 bytes each)
		"Ñoël Café naïve résumé",   // Accented Latin characters
		"ăâîșțĂÂÎȘȚ",               // Romanian diacritics
		"αβγδεζηθικλμνξοπρστυφχψω", // Greek alphabet
		"שלום עולם",                // Hebrew characters (RTL)
		"नमस्कार संसार",            // Hindi characters (Devanagari)
		"สวัสดีชาวโลก",             // Thai characters
		"日本語ひらがなカタカナ漢字",            // Mixed Japanese scripts
		// Complex emojis with modifiers
		"👨‍👩‍👧‍👦",          // Family emoji (ZWJ sequence)
		"👨‍💻👩‍🔬👨‍🍳",        // Profession emojis (ZWJ sequences)
		"👋🏽👋🏾👋🏿",           // Skin tone modifiers
		"🏴󠁧󠁢󠁥󠁮󠁧󠁿🏴󠁧󠁢󠁳󠁣󠁴󠁿",   // Flag emojis with tag sequences
		"👨‍👨‍👧‍👦👩‍👩‍👧‍👧",   // Same-sex family emojis
		"🧑‍🦰🧑‍🦱🧑‍🦳🧑‍🦲",     // Hair style modifiers
		"🏃‍♂️🏃‍♀️🚴‍♂️🚴‍♀️", // Gendered activities
		// Zero-width and combining characters
		"​‌‍‎‏",    // Zero-width characters (ZWSP, ZWNJ, ZWJ, LRM, RLM)
		"e⃝é̲n̸g̃", // Combining diacritical marks
		"à́̂̃̄",   // Multiple combining marks
		"\uFEFF",   // Zero-width no-break space (BOM)
		// Unicode whitespace variants
		" ​ ‌ ‍ ‎ ‏", // Various Unicode spaces
		"　",          // Ideographic space
		// Boundary test cases with multibyte characters
		strings.Repeat("你", MaxTitleLength),     // Exactly max length (Chinese characters)
		strings.Repeat("안", MaxTitleLength),     // Exactly max length (Korean characters)
		strings.Repeat("👨‍💻", MaxTitleLength),   // Exactly max length (ZWJ emoji)
		strings.Repeat("你", MaxTitleLength+1),   // Over max length (Chinese characters)
		strings.Repeat("안", MaxTitleLength+1),   // Over max length (Korean characters)
		strings.Repeat("👨‍💻", MaxTitleLength+1), // Over max length (ZWJ emoji)
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, title string) {
		err := validateTitle(title)

		// Use the same trimming logic as validateTitle function
		trimmed := strings.TrimSpace(title)
		trimmed = strings.TrimFunc(trimmed, func(r rune) bool {
			return unicode.IsSpace(r) ||
				r == '\u200B' || // Zero-width space
				r == '\u200C' || // Zero-width non-joiner
				r == '\u200D' || // Zero-width joiner
				r == '\u200E' || // Left-to-right mark
				r == '\u200F' || // Right-to-left mark
				r == '\uFEFF' // Zero-width no-break space (BOM)
		})
		runeCount := utf8.RuneCountInString(title)

		if trimmed == "" {
			if !errors.Is(err, ErrTitleEmpty) {
				t.Errorf("Expected ErrTitleEmpty for empty/whitespace title %q, got %v", title, err)
			}
		} else if runeCount > MaxTitleLength {
			if !errors.Is(err, ErrTitleTooLong) {
				t.Errorf("Expected ErrTitleTooLong for title %q (runes: %d), got %v", title, runeCount, err)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for valid title %q (runes: %d), got %v", title, runeCount, err)
			}
		}

		byteLen := len(title)
		if runeCount > byteLen {
			t.Errorf("Rune count (%d) should not exceed byte length (%d) for title %q", runeCount, byteLen, title)
		}
	})
}

// FuzzNewTaskWithValidation tests the NewTask function
func FuzzNewTaskWithValidation(f *testing.F) {
	testCases := []string{
		"Valid Task",
		"日本語のタスク",
		"🚀 Rocket Task",
		"",
		"   ",
		strings.Repeat("a", MaxTitleLength),
		strings.Repeat("a", MaxTitleLength+1),
		strings.Repeat("界", MaxTitleLength),
		strings.Repeat("界", MaxTitleLength+1),
		"Mixed 你好世界 Hello World", // Mixed multibyte and ASCII characters
		"Special chars: !@#$%^&*()",
		"\n\t\r",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, title string) {
		taskID := GenerateTaskID()
		userID, err := user.NewUserID("550e8400-e29b-41d4-a716-446655440000")
		if err != nil {
			t.Fatalf("Failed to create test user ID: %v", err)
		}

		task, err := NewTask(taskID, title, userID)

		trimmed := strings.TrimSpace(title)
		trimmed = strings.TrimFunc(trimmed, func(r rune) bool {
			return unicode.IsSpace(r) ||
				r == '\u200B' || // Zero-width space
				r == '\u200C' || // Zero-width non-joiner
				r == '\u200D' || // Zero-width joiner
				r == '\u200E' || // Left-to-right mark
				r == '\u200F' || // Right-to-left mark
				r == '\uFEFF' // Zero-width no-break space (BOM)
		})
		runeCount := utf8.RuneCountInString(title)

		if trimmed == "" {
			// Empty titles should fail
			if err != ErrTitleEmpty {
				t.Errorf("Expected ErrTitleEmpty for empty title %q, got %v", title, err)
			}
			if task != nil {
				t.Errorf("Expected nil task for invalid title %q, got %v", title, task)
			}
		} else if runeCount > MaxTitleLength {
			if !errors.Is(err, ErrTitleTooLong) {
				t.Errorf("Expected ErrTitleTooLong for title %q (runes: %d), got %v", title, runeCount, err)
			}
			if task != nil {
				t.Errorf("Expected nil task for invalid title %q, got %v", title, task)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for valid title %q (runes: %d), got %v", title, runeCount, err)
			}
			if task == nil {
				t.Errorf("Expected non-nil task for valid title %q, got nil", title)
			} else {
				if task.ID() != taskID {
					t.Errorf("Task ID mismatch: expected %v, got %v", taskID, task.ID())
				}
				if task.Title() != title {
					t.Errorf("Task title mismatch: expected %q, got %q", title, task.Title())
				}
				if task.UserID() != userID {
					t.Errorf("Task user ID mismatch: expected %v, got %v", userID, task.UserID())
				}
			}
		}
	})
}
