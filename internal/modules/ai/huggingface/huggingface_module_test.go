package huggingface

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}


// =============================================================================
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local client, err = hf.client({token = "test-token"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingToken(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local client, err = hf.client({})
		-- May work without token for some public models
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// inference tests
// =============================================================================

func TestInferenceNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local result, err = hf.inference(nil, "gpt2", {inputs = "Hello"})
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// text_generation tests
// =============================================================================

func TestTextGenerationNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local text, err = hf.text_generation(nil, "gpt2", "Hello, how are")
		assert(text == nil, "text should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// fill_mask tests
// =============================================================================

func TestFillMaskNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local result, err = hf.fill_mask(nil, "bert-base-uncased", "The capital of France is [MASK].")
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// summarization tests
// =============================================================================

func TestSummarizationNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local summary, err = hf.summarization(nil, "facebook/bart-large-cnn", "Long text here...")
		assert(summary == nil, "summary should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// question_answering tests
// =============================================================================

func TestQuestionAnsweringNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local answer, err = hf.question_answering(nil, "deepset/roberta-base-squad2", {
			question = "What is the capital of France?",
			context = "France is a country in Europe. Paris is its capital."
		})
		assert(answer == nil, "answer should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// translation tests
// =============================================================================

func TestTranslationNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local translated, err = hf.translation(nil, "Helsinki-NLP/opus-mt-en-fr", "Hello world")
		assert(translated == nil, "translated should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// sentiment tests
// =============================================================================

func TestSentimentNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local result, err = hf.sentiment(nil, "distilbert-base-uncased-finetuned-sst-2-english", "I love this!")
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// embeddings tests
// =============================================================================

func TestEmbeddingsNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local embeddings, err = hf.embeddings(nil, "sentence-transformers/all-MiniLM-L6-v2", "Hello world")
		assert(embeddings == nil, "embeddings should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// image_classification tests
// =============================================================================

func TestImageClassificationNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local result, err = hf.image_classification(nil, "google/vit-base-patch16-224", "/path/to/image.jpg")
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// zero_shot tests
// =============================================================================

func TestZeroShotNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local hf = require("ai.huggingface")
		local result, err = hf.zero_shot(nil, "facebook/bart-large-mnli", "This is about cooking", {"food", "sports", "politics"})
		assert(result == nil, "result should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
