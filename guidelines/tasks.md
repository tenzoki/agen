# Phase 4: Cellorg Token Budget Integration

**Status**: Concept / Design Phase
**Date**: 2025-10-08
**Context**: Extending token budget management from alfa (AI workbench) to cellorg (agent orchestration)

---

## Overview

Phase 1-3 implemented token budget management for alfa's direct AI interactions. Phase 4 extends this capability to **agent-to-agent communication** within cellorg, enabling agents to automatically handle large payloads that exceed token limits when communicating with AI-powered agents.

## Problem Statement

Currently, when an agent sends a large document/dataset to an AI-powered agent (e.g., text_analyzer, summary_generator, embedding_agent):

1. **No automatic chunking**: Large payloads fail with "max context exceeded" errors
2. **No response merging**: Agents can't handle streaming multi-chunk responses
3. **No budget awareness**: Agents don't know if their payload fits within limits

This breaks workflows like:

- Large document analysis pipelines
- Bulk text processing cells
- Multi-document RAG workflows

## Architecture Principles Compliance

✅ **Cells-First Design**: Chunking happens at envelope level, transparent to agents
✅ **Zero-Boilerplate Agents**: Framework handles complexity, agents stay simple
✅ **Public API Only**: Uses existing envelope/broker infrastructure
✅ **No Breaking Changes**: Backward compatible with existing agents

---

## Design

### 1. Token Budget Manager in Envelope Processing

**Location**: `cellorg/internal/envelope/budget.go`

Add token budget calculation to envelope creation/routing:

```go
type EnvelopeBudget struct {
    // From omni/budget
    Manager *budget.Manager

    // Envelope-specific
    PayloadTokens   int
    HeaderTokens    int
    TotalTokens     int
    NeedsSplitting  bool
    SuggestedChunks int
}

// CalculateBudget estimates token usage for an envelope
func CalculateBudget(env *Envelope, counter tokencount.Counter) (*EnvelopeBudget, error) {
    // Count payload tokens
    payloadStr := string(env.Payload)
    payloadTokens, _ := counter.Count(payloadStr)

    // Count header/metadata tokens (conservative)
    headerTokens := estimateMetadataTokens(env)

    return &EnvelopeBudget{
        PayloadTokens: payloadTokens,
        HeaderTokens:  headerTokens,
        TotalTokens:   payloadTokens + headerTokens,
        // ... budget calculation
    }
}
```

### 2. Chunked Envelope Support

**Location**: `cellorg/internal/envelope/chunking.go`

Add envelope splitting for large payloads:

```go
// ChunkEnvelope splits a large envelope into manageable chunks
func ChunkEnvelope(env *Envelope, budget *EnvelopeBudget) ([]*Envelope, error) {
    if !budget.NeedsSplitting {
        return []*Envelope{env}, nil
    }

    // Parse payload as JSON or text
    var chunks []string
    if isJSONArray(env.Payload) {
        chunks = splitJSONArray(env.Payload, budget.SuggestedChunks)
    } else {
        chunks = splitTextPayload(env.Payload, budget.SuggestedChunks)
    }

    // Create chunk envelopes
    envelopes := make([]*Envelope, len(chunks))
    chunkID := uuid.New().String() // Group ID for all chunks

    for i, chunk := range chunks {
        envelopes[i] = &Envelope{
            ID:          uuid.New().String(),
            CorrelationID: env.ID, // Link to original
            Source:      env.Source,
            Destination: env.Destination,
            MessageType: env.MessageType,
            Payload:     json.RawMessage(chunk),

            // Chunk metadata in headers
            Headers: map[string]string{
                "X-Chunk-ID":    chunkID,
                "X-Chunk-Index": strconv.Itoa(i),
                "X-Chunk-Total": strconv.Itoa(len(chunks)),
            },

            TraceID:  env.TraceID,
            SpanID:   uuid.New().String(),
        }
    }

    return envelopes, nil
}
```

### 3. Response Merger for Agents

**Location**: `cellorg/public/agent/chunking.go`

Public API for agents to handle chunked responses:

```go
type ChunkCollector struct {
    chunks    map[string][]*Envelope // chunkID -> ordered chunks
    mu        sync.Mutex
    waitChans map[string]chan *Envelope // chunkID -> completion channel
}

// CollectChunk accumulates chunks and returns complete message when ready
func (cc *ChunkCollector) CollectChunk(env *Envelope) (*Envelope, bool, error) {
    chunkID := env.Headers["X-Chunk-ID"]
    if chunkID == "" {
        // Not a chunked message, return immediately
        return env, true, nil
    }

    cc.mu.Lock()
    defer cc.mu.Unlock()

    // Accumulate chunk
    if cc.chunks[chunkID] == nil {
        cc.chunks[chunkID] = make([]*Envelope, 0)
    }
    cc.chunks[chunkID] = append(cc.chunks[chunkID], env)

    // Check if complete
    totalChunks, _ := strconv.Atoi(env.Headers["X-Chunk-Total"])
    if len(cc.chunks[chunkID]) == totalChunks {
        merged := mergeChunks(cc.chunks[chunkID])
        delete(cc.chunks, chunkID)
        return merged, true, nil // Complete!
    }

    return nil, false, nil // Still waiting for more chunks
}

// mergeChunks combines chunk payloads
func mergeChunks(chunks []*Envelope) *Envelope {
    // Sort by chunk index
    sort.Slice(chunks, func(i, j int) bool {
        idxI, _ := strconv.Atoi(chunks[i].Headers["X-Chunk-Index"])
        idxJ, _ := strconv.Atoi(chunks[j].Headers["X-Chunk-Index"])
        return idxI < idxJ
    })

    // Merge payloads
    var merged []byte
    for _, chunk := range chunks {
        merged = append(merged, chunk.Payload...)
    }

    // Create merged envelope (use first chunk as template)
    result := *chunks[0]
    result.Payload = json.RawMessage(merged)
    delete(result.Headers, "X-Chunk-ID")
    delete(result.Headers, "X-Chunk-Index")
    delete(result.Headers, "X-Chunk-Total")

    return &result
}
```

### 4. Broker Integration

**Location**: `cellorg/internal/broker/chunking.go`

Broker automatically chunks large messages before routing:

```go
type ChunkingBroker struct {
    baseBroker *Broker
    counters   map[string]tokencount.Counter // provider -> counter
}

// PublishWithChunking checks budget and chunks if needed
func (cb *ChunkingBroker) PublishWithChunking(env *Envelope) error {
    // Determine target agent's AI provider (from config)
    targetProvider := cb.getTargetProvider(env.Destination)
    if targetProvider == "" {
        // Non-AI agent, no chunking needed
        return cb.baseBroker.Publish(env)
    }

    // Get token counter for target provider
    counter := cb.counters[targetProvider]
    if counter == nil {
        return cb.baseBroker.Publish(env) // Fallback
    }

    // Calculate budget
    budget, err := CalculateBudget(env, counter)
    if err != nil {
        return cb.baseBroker.Publish(env) // Fallback
    }

    if !budget.NeedsSplitting {
        return cb.baseBroker.Publish(env)
    }

    // Split into chunks
    chunks, err := ChunkEnvelope(env, budget)
    if err != nil {
        return err
    }

    // Publish all chunks
    for _, chunk := range chunks {
        if err := cb.baseBroker.Publish(chunk); err != nil {
            return err
        }
    }

    return nil
}
```

### 5. Agent Framework Integration

**Location**: `cellorg/public/agent/framework.go`

Update framework to provide chunking support:

```go
type AgentFramework struct {
    // ... existing fields
    chunkCollector *ChunkCollector // For receiving chunked messages
}

// In startMessageProcessing():
func (f *AgentFramework) processMessage(msg *client.BrokerMessage) error {
    env := &envelope.Envelope{...} // Convert BrokerMessage to Envelope

    // Handle chunked messages
    mergedEnv, complete, err := f.chunkCollector.CollectChunk(env)
    if err != nil {
        return err
    }
    if !complete {
        return nil // Still waiting for more chunks
    }

    // Process complete message
    return f.runner.ProcessMessage(f.baseAgent, mergedEnv)
}
```

---

## Implementation Plan

### Step 1: Core Budget Integration

**Files**:

- `cellorg/internal/envelope/budget.go` (new)
- `cellorg/go.mod` (add omni/tokencount dependency)

**Tasks**:

- Add `CalculateBudget()` function
- Add `estimateMetadataTokens()` helper
- Write tests for envelope token counting

### Step 2: Chunking Support

**Files**:

- `cellorg/internal/envelope/chunking.go` (new)
- `cellorg/internal/envelope/chunking_test.go` (new)

**Tasks**:

- Implement `ChunkEnvelope()` with JSON and text splitting
- Add chunk metadata headers
- Test chunk ordering and merging

### Step 3: Agent-Side Chunk Collection

**Files**:

- `cellorg/public/agent/chunking.go` (new)
- `cellorg/public/agent/chunking_test.go` (new)

**Tasks**:

- Implement `ChunkCollector` for receiving chunks
- Implement `mergeChunks()` with ordering
- Add timeout handling for incomplete chunks

### Step 4: Broker Integration

**Files**:

- `cellorg/internal/broker/chunking.go` (new)
- `cellorg/internal/broker/service.go` (update)

**Tasks**:

- Add `ChunkingBroker` wrapper
- Implement `PublishWithChunking()`
- Add provider detection from agent config

### Step 5: Framework Integration

**Files**:

- `cellorg/public/agent/framework.go` (update)
- `cellorg/public/agent/framework_test.go` (update)

**Tasks**:

- Add `ChunkCollector` to framework
- Update message processing loop
- Handle chunk timeout and cleanup

### Step 6: Configuration

**Files**:

- `workbench/config/pool.yaml` (update schema)
- `workbench/config/cells/*.yaml` (examples)

**Tasks**:

- Add token budget config to agent pool
- Document chunking behavior
- Provide example configurations

---

## Configuration Schema

Add to agent pool entries:

```yaml
agents:
  summary_generator:
    type: summary_generator
    ai_provider: anthropic
    ai_model: claude-sonnet-4-5-20250929
    token_budget:
      enabled: true
      safety_margin: 0.10
      max_chunk_size: 150000  # tokens per chunk
      chunk_overlap: 0.15     # 15% overlap for context
```

Add to cell configurations:

```yaml
cells:
  document_analysis:
    chunking:
      enabled: true
      strategy: auto  # auto, json, text
      max_payload_size: 1000000  # bytes
```

---

## Backward Compatibility

✅ **No Breaking Changes**:

- Chunking is opt-in via configuration
- Non-chunked envelopes work as before
- Agents without chunk support work normally

✅ **Graceful Degradation**:

- If token counting fails → send without chunking
- If budget calculation errors → fallback to normal flow
- Missing chunk timeouts handled gracefully

---

## Testing Strategy

### Unit Tests

- Envelope token counting accuracy
- Chunk splitting with various payload types
- Chunk ordering and merging
- Metadata preservation across chunks

### Integration Tests

- End-to-end chunked message flow
- Multi-agent pipeline with large payloads
- Timeout handling for incomplete chunks
- Mixed chunked/non-chunked messages

### Performance Tests

- Overhead of token counting
- Chunking/merging latency
- Memory usage with many concurrent chunks

---

## Benefits

1. **Automatic Handling**: Agents automatically handle large payloads
2. **Transparent**: Works without agent code changes
3. **Provider-Aware**: Uses correct token limits per AI provider
4. **Fault Tolerant**: Graceful fallback if chunking fails
5. **Observable**: Chunk IDs enable distributed tracing

---

## Open Questions

1. **Chunk Timeout**: How long to wait for incomplete chunks? (Propose: 5 minutes)
2. **Chunk Storage**: Keep chunks in memory or persist? (Propose: memory with size limits)
3. **Retry Logic**: Retry sending if chunk delivery fails? (Propose: yes, with exponential backoff)
4. **Deduplication**: Handle duplicate chunks? (Propose: yes, use chunk index)
5. **Compression**: Compress chunks before sending? (Propose: optional, add later)

---

## Risks & Mitigations

**Risk**: Chunking adds latency
**Mitigation**: Make opt-in, only for AI-powered agents with large payloads

**Risk**: Chunk reassembly memory usage
**Mitigation**: Set max concurrent chunks limit, timeout cleanup

**Risk**: Chunk ordering issues
**Mitigation**: Use explicit chunk index in headers, validate on merge

**Risk**: Incomplete chunks blocking agents
**Mitigation**: Timeout mechanism, log warnings, continue processing

---

## Next Steps

1. Review this concept with human collaborator
2. Get approval for architecture approach
3. Begin Step 1: Core Budget Integration
4. Implement incrementally with tests
5. Update documentation as we progress

---

## References

- Phase 1-3 implementation: `code/omni/tokencount/`, `code/omni/budget/`, `code/alfa/internal/orchestrator/`
- Envelope structure: `code/cellorg/internal/envelope/envelope.go`
- Agent framework: `code/cellorg/public/agent/framework.go`
- Broker service: `code/cellorg/internal/broker/service.go`
