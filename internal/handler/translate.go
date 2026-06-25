//nolint:goconst // JSON response field names are domain strings, not config constants
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// TranslateRequest is the request body for POST /api/translate.
type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

// TranslateResponse is the response for POST /api/translate.
type TranslateResponse struct {
	TranslatedText string `json:"translated_text"`
	SourceLang     string `json:"source_lang"`
	TargetLang     string `json:"target_lang"`
}

// ServeTranslate handles /api/translate requests.
func ServeTranslate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeLocalizedErrorf(w, r, http.StatusMethodNotAllowed, "MethodNotAllowed")
		return
	}

	var req TranslateRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Text == "" {
		writeLocalizedErrorf(w, r, http.StatusBadRequest, "InvalidRequestBody")
		return
	}

	// Default languages
	if req.SourceLang == "" {
		req.SourceLang = "en"
	}
	if req.TargetLang == "" {
		req.TargetLang = "zh"
	}

	// Translate using MyMemory
	translatedText, err := translateWithMyMemory(req.Text, req.SourceLang, req.TargetLang)
	if err != nil {
		slog.Error("translation failed", "error", err)
		writeLocalizedErrorf(w, r, http.StatusInternalServerError, "TranslationFailed")
		return
	}

	writeJSON(w, http.StatusOK, TranslateResponse{
		TranslatedText: translatedText,
		SourceLang:     req.SourceLang,
		TargetLang:     req.TargetLang,
	})
}

// translateWithMyMemory translates text using MyMemory free API.
// Splits long text into chunks of 500 characters and translates each chunk.
func translateWithMyMemory(text, sourceLang, targetLang string) (string, error) {
	const maxQueryLength = 500

	// If text is short enough, translate directly
	if len(text) <= maxQueryLength {
		return translateChunk(text, sourceLang, targetLang)
	}

	// Split text into chunks
	chunks := splitTextIntoChunks(text, maxQueryLength)
	var translatedChunks []string

	for _, chunk := range chunks {
		translated, err := translateChunk(chunk, sourceLang, targetLang)
		if err != nil {
			return "", fmt.Errorf("failed to translate chunk: %w", err)
		}
		translatedChunks = append(translatedChunks, translated)
	}

	// Join translated chunks
	return joinChunks(translatedChunks), nil
}

// translateChunk translates a single chunk of text.
func translateChunk(text, sourceLang, targetLang string) (string, error) {
	// Build URL
	baseURL := "https://api.mymemory.translated.net/get"
	params := url.Values{
		"q":       {text},
		"langpair": {sourceLang + "|" + targetLang},
	}
	translateURL := baseURL + "?" + params.Encode()

	// Create request with custom User-Agent
	req, err := http.NewRequest("GET", translateURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("translation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	// Response format: {"responseData":{"translatedText":"translated text"},...}
	slog.Info("translation response", "body", string(body))
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract translated text
	if responseData, ok := result["responseData"].(map[string]interface{}); ok {
		if translatedText, ok := responseData["translatedText"].(string); ok {
			return translatedText, nil
		}
	}

	return "", fmt.Errorf("unexpected response format")
}

// splitTextIntoChunks splits text into chunks of maxLen characters.
// Tries to split at sentence boundaries when possible.
func splitTextIntoChunks(text string, maxLen int) []string {
	var chunks []string
	runes := []rune(text)
	totalLen := len(runes)

	for i := 0; i < totalLen; {
		end := i + maxLen
		if end > totalLen {
			end = totalLen
		}

		// Try to find a good split point (sentence boundary)
		if end < totalLen {
			// Look for sentence endings
			for j := end - 1; j > i+maxLen/2; j-- {
				c := runes[j]
				if c == '.' || c == '!' || c == '?' || c == ';' || c == '\n' {
					end = j + 1
					break
				}
			}
		}

		chunk := string(runes[i:end])
		chunks = append(chunks, chunk)
		i = end
	}

	return chunks
}

// joinChunks joins translated chunks back together.
func joinChunks(chunks []string) string {
	var result string
	for i, chunk := range chunks {
		if i > 0 {
			// Add space between chunks if needed
			if len(result) > 0 && len(chunk) > 0 {
				lastChar := rune(result[len(result)-1])
				firstChar := rune(chunk[0])
				if lastChar != ' ' && firstChar != ' ' {
					result += " "
				}
			}
		}
		result += chunk
	}
	return result
}
