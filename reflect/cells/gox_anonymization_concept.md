# GOX Anonymization Concept (Finalized)

## 🎯 Goal
Implement fully offline, reversible anonymization of sensitive entities in multilingual text:
- Persons, companies, institutions, products, places
- Support **all major European languages**, including German, English, French, Spanish, Slavic, and more
- Ensure **consistent pseudonyms**, **reversibility**, and **secure local storage**
- Avoid weak MVPs (no simplified NER-only version)

---

## 🧩 Core Pipeline

### 1. NER (Named Entity Recognition)
- **Technology**: ONNXRuntime with HuggingFace multilingual models
- **Models**:
  - [`xlm-roberta-large-finetuned-conll03`](https://huggingface.co/xlm-roberta-large-finetuned-conll03-english)
  - [`Davlan/xlm-roberta-base-ner-hrl`](https://huggingface.co/Davlan/xlm-roberta-base-ner-hrl)
- **Purpose**: Extract entity mentions (PERSON, ORG, LOC, PRODUCT)
- **Reason**: High multilingual accuracy, fully offline, production-ready

### 2. Coreference Resolution
- **Technology**: ONNXRuntime, SpanBERT-based models
- **Model**: [`allenai/coref-spanbert-large`](https://huggingface.co/allenai/coref-spanbert-large)
- **Purpose**: Cluster mentions that refer to the same entity  
  Example: *“Angela Merkel … she … the Chancellor”* → unified to one entity

### 3. Synonym & Variant Normalization
- **Technology**: Embedding similarity using multilingual sentence transformers
- **Model**: [`sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`](https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2)
- **Purpose**: Merge synonyms, nicknames, abbreviations, spelling variants  
  Example: *“Robert”, “Bob”* → one canonical entity

### 4. Anonymization & Reversibility
- **Pseudonym generation**:
  - Prefix by entity type (`PERSON_123`, `ORG_456`, `LOC_789`)
  - Deterministic ID (hash-based seed for stability)
- **Mapping store**:
  - Uses Gox's existing storage infrastructure (`godast/omnistore` with bbolt)
  - Bidirectional mapping (original ↔ pseudonym)
  - Storage agent pattern: `anonymization_store` agent
- **Security**:
  - AES encryption for mapping store
  - Timestamp, pipeline ID, model version stored for audit

---

## 📐 Architecture

```go
type Entity struct {
    Text     string
    Type     string  // PERSON, ORG, LOC, PRODUCT
    Start    int
    End      int
    Canonical string // normalized form
}

type Anonymizer struct {
    ner          NEREngine
    coref        CorefEngine
    linker       SynonymLinker
    store        MappingStore
}

type MappingStore interface {
    Store(original, canonical, pseudonym, entityType string) error
    LookupOriginal(pseudonym string) (string, error)
    LookupPseudonym(original string) (string, error)
}
```

---

## 🔄 Workflow

```ascii
Input Text
   │
   ▼
[NER] ──▶ Entities
   │
   ▼
[Coreference Resolution] ──▶ Cluster mentions
   │
   ▼
[Synonym/Embedding Linking] ──▶ Canonical entities
   │
   ▼
[Anonymizer] ──▶ Pseudonymized Text + Mapping Store
   │
   ▼
(Reversible via Store)
```

---

## 🚀 Implementation Notes
- **ONNXRuntime-Go** for all models (NER, Coref, Embeddings)
- **Gox storage infrastructure** (`godast/omnistore` with bbolt backend)
  - Create `anonymization_store` agent similar to `godast_storage`
  - Agents use `internal/storage.Client` for broker-based storage access
- **AES-GCM** to encrypt store on disk
- Deterministic pseudonym generation:
  ```go
  func pseudonym(entityType, canonical string) string {
      id := hash(canonical) % 1e6
      return fmt.Sprintf("%s_%06d", entityType, id)
  }
  ```

---

## ✅ Advantages
- **Multilingual**: Robust across European languages
- **Offline**: No data leaves GoX, cloud-safe
- **Accurate**: Modern transformer-based NER + coref + embeddings
- **Consistent**: Same entity → same pseudonym across documents
- **Reversible**: Secure mapping enables controlled deanonymization
- **Auditable**: Metadata persisted for compliance

---
