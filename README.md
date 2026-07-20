# go-federated-registry
go-federated-registry

## Helper

```sh
tr -d '[],' < v1.txt | xargs
tr -d '[],' < v2.txt | xargs | tr ' ' ','
```

## Diagram

```mermaid
sequenceDiagram
    autonumber
    actor Client as Orchestrator / Multi-Agent System
    participant Service as go-federated-registry
    participant LLM as TEI Hugging Face (embedding vector)
    participant DB as DB pgvector
    
    Client->>Service: "Find me a service that can fetch inventory data"
    Service->>LLM: Generate Embedding for Query String
    LLM-->>Service: Return Vector [0.012, -0.043, ...]
    Service->>DB: Execute Hybrid SQL Query (Vector Distance + Health Filter)
    Note over DB: Computes Cosine Distance<br/>Filters WHERE health_status = 'HEALTHY'
    DB-->>Service: Return service record (sales-mcp)
    Service-->>Client: Return Base URL & Execution Endpoints
```

```mermaid
sequenceDiagram
    autonumber
    participant Engine as Query Planner
    participant Filter as Relational Filter
    participant Index as HNSW Vector Index
    participant Project as Projection Selector

    Engine->>Filter: 1. Join tables & filter on WHERE health_status = 'HEALTHY'
    Filter-->>Engine: Filtered Candidates
    Engine->>Index: 2. Traversed index using <=> operator
    Index-->>Engine: Sorted Row IDs (by Distance ASC)
    Engine->>Project: 3. Fetch SELECT columns & calculate score
    Project-->>Engine: Final 5 Rows (LIMIT 5)

```

## To do

1. Evaluate a HNSW index

select 	sr.id,
		sr.base_transport, 
		sr.health_status,
		se.url
from 	service_registry sr,
		service_endpoints se,
		service_vectors sv
where 	se.fk_service_registry_id = sr.id
and 	sv.search_vector <=> '[-0.010023851,0.0028419404,...]'
limit 5; 
