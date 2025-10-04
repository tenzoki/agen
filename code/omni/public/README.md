# Public

Public API for OmniStore.

## Intent

Defines the public interface and types that external packages can use to interact with OmniStore. Contains only the OmniStore interface and configuration types.

## Usage

```go
import "agen/code/omni/public/omnistore"

store, err := omnistore.NewOmniStoreWithDefaults("/data")
```

## Setup

No dependencies beyond the parent module.

## Tests

```bash
go test ./public/omnistore
```

## Demo

See parent README for demo references.
