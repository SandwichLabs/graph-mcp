
Agent Memory Graph MCP Project
============================

## Executive Summary


This document provides a comprehensive implementation guide and Domain Driven Design (DDD) specification for building a sophisticated memory graph system using KuzuDB, Go, and the Model Context Protocol (MCP). The system will support both Terminal User Interface (TUI) via Charm's Bubble Tea framework and MCP server capabilities for LLM integration.


## 1. Domain Driven Design Specification


### 1.1 Domain Model Overview


The memory graph domain centers around the concept of a **Knowledge Repository** that stores and retrieves interconnected information through both graph traversal and vector similarity search.


### 1.2 Core Domain Entities


#### 1.2.1 Aggregates


**Knowledge Graph Aggregate**
- **Root Entity**: `Graph`
- **Value Objects**: `GraphMetadata`, `SchemaVersion`
- **Entities**: `Source`, `Chunk`, `Entity`, `Concept`
- **Invariants**: 
  - All chunks must belong to a source
  - All relationships must have valid source and target nodes
  - Vector embeddings must match configured dimensions


**Query Aggregate**
- **Root Entity**: `Query`
- **Value Objects**: `QueryResult`, `QueryMetrics`
- **Entities**: `ExecutionPlan`, `ResultSet`


**Ingestion Aggregate**
- **Root Entity**: `IngestionJob`
- **Value Objects**: `IngestionStatus`, `ChunkingStrategy`
- **Entities**: `Document`, `ExtractedTriple`


#### 1.2.2 Value Objects


```go
// Domain value objects
type NodeID struct {
    Value uuid.UUID
}


type EmbeddingVector struct {
    Dimensions int
    Values     []float32
}


type Predicate struct {
    Value string
}


type ConfidenceScore struct {
    Value float32 // 0.0 to 1.0
}


type ChunkMetadata struct {
    Index      int
    StartChar  int
    EndChar    int
    Overlap    int
}
```


#### 1.2.3 Domain Events


```go
// Event definitions
type SourceIngested struct {
    SourceID   NodeID
    URI        string
    Timestamp  time.Time
}


type EntityDiscovered struct {
    EntityID   NodeID
    Name       string
    Type       string
    SourceID   NodeID
}


type RelationshipEstablished struct {
    SubjectID  NodeID
    Predicate  Predicate
    ObjectID   NodeID
    Confidence ConfidenceScore
}


type QueryExecuted struct {
    QueryID    uuid.UUID
    Query      string
    ResultCount int
    Duration   time.Duration
}
```


### 1.3 Bounded Contexts


1. **Knowledge Management Context**
   - Handles graph schema, node/edge creation, updates
   - Owns: Graph, Source, Chunk, Entity, Concept


2. **Query Processing Context**
   - Handles Cypher query execution, vector search
   - Owns: Query, QueryResult, ExecutionPlan


3. **Ingestion Context**
   - Handles document processing, LLM extraction
   - Owns: IngestionJob, Document, ExtractedTriple


4. **User Interface Context**
   - Handles TUI and MCP interactions
   - Owns: UIState, MCPRequest, MCPResponse


### 1.4 Anti-Corruption Layers


- **LLM Provider Adapter**: Abstracts OpenAI, Ollama, etc.
- **Database Adapter**: Wraps KuzuDB-specific operations
- **File System Adapter**: Handles various document formats


## 2. Technical Architecture


### 2.1 System Architecture Diagram


```
┌─────────────────────────────────────────────────────────────────┐
│                         User Layer                               │
├─────────────────────────┬───────────────────────────────────────┤
│     TUI (Bubble Tea)    │        MCP Server (JSON-RPC)         │
├─────────────────────────┴───────────────────────────────────────┤
│                    Application Layer                             │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │   Command   │  │    Query     │  │    Ingestion      │    │
│  │   Handlers  │  │   Service    │  │     Service       │    │
│  └─────────────┘  └──────────────┘  └────────────────────┘    │
├─────────────────────────────────────────────────────────────────┤
│                     Domain Layer                                 │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │   Graph     │  │   Entity     │  │   Relationship    │    │
│  │   Model     │  │ Repository   │  │   Repository      │    │
│  └─────────────┘  └──────────────┘  └────────────────────┘    │
├─────────────────────────────────────────────────────────────────┤
│                 Infrastructure Layer                             │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │   KuzuDB    │  │     LLM      │  │   File System     │    │
│  │   Adapter   │  │   Adapter    │  │     Adapter       │    │
│  └─────────────┘  └──────────────┘  └────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```


### 2.2 Project Structure


```
/memgraph
├── /cmd
│   ├── /memgraph         # Main CLI entry point
│   │   └── main.go
│   └── /memgraph-server  # MCP server entry point
│       └── main.go
├── /internal
│   ├── /domain           # Core domain logic
│   │   ├── /graph        # Graph entities and value objects
│   │   ├── /query        # Query models
│   │   └── /ingestion    # Ingestion models
│   ├── /application      # Application services
│   │   ├── /commands     # Command handlers
│   │   ├── /queries      # Query handlers
│   │   └── /services     # Domain services
│   ├── /infrastructure   # External adapters
│   │   ├── /kuzu         # KuzuDB implementation
│   │   ├── /llm          # LLM providers
│   │   └── /storage      # File system operations
│   ├── /interfaces       # User interfaces
│   │   ├── /tui          # Bubble Tea TUI
│   │   ├── /mcp          # MCP server handlers
│   │   └── /cli          # Cobra CLI commands
│   └── /shared           # Shared utilities
│       ├── /errors       # Domain errors
│       └── /logging      # Structured logging
├── /pkg                  # Public packages
│   └── /memgraph         # Public API
├── /configs              # Configuration files
├── /scripts              # Build and deployment scripts
├── /docs                 # Documentation
├── go.mod
├── go.sum
├── Makefile
└── README.md
```


### 2.3 Key Technical Components


#### 2.3.1 CLI Tool Specification (Unix-like)


```bash
# Core commands following Unix philosophy
memgraph init [path]                    # Initialize new graph database
memgraph ingest <source>                # Ingest document/URL
memgraph query <cypher>                 # Execute Cypher query
memgraph search <text>                  # Vector similarity search
memgraph add <subject> <predicate> <object>  # Add fact
memgraph ls [pattern]                   # List entities
memgraph show <entity>                  # Show entity details
memgraph export <format>                # Export graph
memgraph import <file>                  # Import graph


# Modifiers (Unix-style flags)
--config, -c <file>    # Configuration file
--output, -o <format>  # Output format (json|table|csv)
--limit, -l <n>        # Limit results
--verbose, -v          # Verbose output
--quiet, -q            # Quiet mode
--help, -h             # Show help


# Pipe-friendly operations
memgraph query "MATCH (n) RETURN n" | jq '.nodes[]'
memgraph ls "Person" | grep "Fink" | memgraph show -
```


#### 2.3.2 TUI Interface Components (Bubble Tea)


```go
// TUI Model structure
type Model struct {
    // Views
    currentView  ViewType
    graphView    *GraphViewModel
    queryView    *QueryViewModel
    entityView   *EntityViewModel
    
    // State
    graph        *domain.Graph
    queryHistory []domain.Query
    
    // UI Components
    viewport     viewport.Model
    textInput    textinput.Model
    table        table.Model
    spinner      spinner.Model
    help         help.Model
}


// View types
type ViewType int
const (
    GraphOverview ViewType = iota
    QueryInterface
    EntityExplorer
    IngestionWizard
    Settings
)
```


#### 2.3.3 MCP Server Specification


```go
// MCP Tool definitions
type MCPTools struct {
    ExecuteQuery    Tool // Read-only Cypher queries
    VectorSearch    Tool // Semantic search
    AddFact         Tool // Add triple
    UpdateEntity    Tool // Update properties
    DeleteRelation  Tool // Remove relationship
    BulkIngest      Tool // Batch operations
}


// MCP Resources
type MCPResources struct {
    GraphSchema     Resource // schema://graph
    EntityDetails   Resource // entity://{name}
    Statistics      Resource // stats://overview
    QueryHistory    Resource // history://queries
}
```


## 3. Implementation Roadmap


### Phase 1: Foundation (Weeks 1-4)


**Sprint 1: Core Infrastructure**
- [ ] Project setup with Go modules
- [ ] KuzuDB integration layer
- [ ] Basic domain models
- [ ] Unit test framework


**Sprint 2: CLI Framework**
- [ ] Cobra CLI structure
- [ ] Basic commands (init, query)
- [ ] Configuration management
- [ ] Error handling patterns


**Deliverables:**
- Working CLI with basic graph operations
- KuzuDB adapter with connection pooling
- Comprehensive test suite


### Phase 2: Knowledge Management (Weeks 5-8)


**Sprint 3: Ingestion Pipeline**
- [ ] Document chunking service
- [ ] LLM extraction interface
- [ ] Bulk loading via COPY FROM
- [ ] Progress reporting


**Sprint 4: Query Engine**
- [ ] Cypher query executor
- [ ] Vector search implementation
- [ ] Hybrid query orchestration
- [ ] Result formatting


**Deliverables:**
- Full ingestion pipeline
- Hybrid search capabilities
- Performance benchmarks


### Phase 3: User Interfaces (Weeks 9-12)


**Sprint 5: TUI Development**
- [ ] Bubble Tea app structure
- [ ] Graph visualization view
- [ ] Query interface with history
- [ ] Entity explorer


**Sprint 6: MCP Server**
- [ ] MCP server setup
- [ ] Tool implementations
- [ ] Resource handlers
- [ ] Authentication/authorization


**Deliverables:**
- Feature-complete TUI
- MCP server with all tools
- Integration tests


### Phase 4: Advanced Features (Weeks 13-16)


**Sprint 7: Intelligence Layer**
- [ ] Router agent pattern
- [ ] Query optimization
- [ ] Metacognitive logging
- [ ] Performance analytics


**Sprint 8: Production Readiness**
- [ ] Docker containerization
- [ ] Deployment scripts
- [ ] Monitoring/observability
- [ ] Documentation


**Deliverables:**
- Production-ready system
- Deployment guide
- Performance report


## 4. Key Implementation Artifacts


### 4.1 Domain Events Implementation


```go
// internal/domain/events/events.go
package events


import (
    "time"
    "github.com/google/uuid"
)


type EventType string


const (
    SourceIngestedEvent EventType = "source.ingested"
    EntityDiscoveredEvent EventType = "entity.discovered"
    RelationshipEstablishedEvent EventType = "relationship.established"
)


type DomainEvent interface {
    GetID() uuid.UUID
    GetType() EventType
    GetTimestamp() time.Time
    GetAggregateID() uuid.UUID
}


type EventBus interface {
    Publish(event DomainEvent) error
    Subscribe(eventType EventType, handler EventHandler) error
}


type EventHandler func(event DomainEvent) error
```


### 4.2 Repository Interfaces


```go
// internal/domain/graph/repository.go
package graph


import (
    "context"
)


type GraphRepository interface {
    // Node operations
    CreateNode(ctx context.Context, node Node) error
    GetNode(ctx context.Context, id NodeID) (*Node, error)
    UpdateNode(ctx context.Context, node Node) error
    DeleteNode(ctx context.Context, id NodeID) error
    
    // Relationship operations
    CreateRelationship(ctx context.Context, rel Relationship) error
    GetRelationships(ctx context.Context, nodeID NodeID) ([]Relationship, error)
    
    // Query operations
    ExecuteCypher(ctx context.Context, query string, params map[string]interface{}) (*ResultSet, error)
    VectorSearch(ctx context.Context, embedding []float32, limit int) ([]Chunk, error)
}
```


### 4.3 Service Layer


```go
// internal/application/services/graph_service.go
package services


type GraphService struct {
    repo      graph.GraphRepository
    llm       llm.Provider
    eventBus  events.EventBus
}


func (s *GraphService) IngestDocument(ctx context.Context, source io.Reader) error {
    // 1. Chunk document
    // 2. Extract entities/relationships via LLM
    // 3. Generate embeddings
    // 4. Store in graph
    // 5. Publish events
}


func (s *GraphService) HybridSearch(ctx context.Context, query string) (*SearchResult, error) {
    // 1. Try Cypher query
    // 2. Fallback to vector search
    // 3. Combine results
    // 4. Return ranked results
}
```


### 4.4 TUI Views


```go
// internal/interfaces/tui/views/graph_view.go
package views


import (
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
)


type GraphView struct {
    viewport viewport.Model
    nodes    []NodeDisplay
    edges    []EdgeDisplay
    selected int
}


func (v GraphView) Update(msg tea.Msg) (GraphView, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "j", "down":
            v.selected++
        case "k", "up":
            v.selected--
        case "enter":
            return v, v.selectNode(v.nodes[v.selected])
        }
    }
    return v, nil
}
```


## 5. Quality Assurance & Testing Strategy


### 5.1 Testing Pyramid


1. **Unit Tests** (70%)
   - Domain logic
   - Service methods
   - Utility functions


2. **Integration Tests** (20%)
   - KuzuDB operations
   - LLM integrations
   - File system operations


3. **E2E Tests** (10%)
   - CLI commands
   - TUI workflows
   - MCP protocols


### 5.2 Performance Benchmarks


```go
// Benchmark targets
type PerformanceTargets struct {
    IngestionRate   int     // 1000 nodes/second
    QueryLatency    time.Duration // <100ms for 1-hop
    VectorSearch    time.Duration // <50ms for top-10
    ConcurrentUsers int     // 100 simultaneous
}
```


## 6. Deployment & Operations


### 6.1 Docker Configuration


```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc g++ make
WORKDIR /app
COPY . .
RUN make build


FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/memgraph /usr/local/bin/
VOLUME ["/data"]
EXPOSE 8080
CMD ["memgraph", "server"]
```


### 6.2 Monitoring & Observability


- **Metrics**: Prometheus integration
- **Logging**: Structured logs with zerolog
- **Tracing**: OpenTelemetry support
- **Health Checks**: HTTP/gRPC endpoints


## 7. Risk Management


### 7.1 Technical Risks


| Risk | Impact | Mitigation |
|------|--------|------------|
| CGo complexity | High | Comprehensive testing, CI/CD pipeline |
| KuzuDB API changes | Medium | Version pinning, adapter pattern |
| LLM extraction quality | Medium | Validation layer, confidence scoring |
| Concurrent access | High | Connection pooling, transaction management |


### 7.2 Mitigation Strategies


1. **Abstraction Layers**: Isolate external dependencies
2. **Feature Flags**: Gradual rollout of new features
3. **Backup Strategy**: Regular graph snapshots
4. **Rollback Plan**: Version-tagged releases


## 8. Success Metrics


### 8.1 Technical KPIs


- Query response time < 100ms (p95)
- Ingestion throughput > 1000 entities/second
- System uptime > 99.9%
- Test coverage > 80%


### 8.2 Business KPIs


- Time to first value < 5 minutes
- User adoption rate > 70%
- Query success rate > 95%
- Documentation completeness 100%


## 9. Team Structure & Responsibilities


### 9.1 Recommended Team Composition


- **Technical Lead**: Architecture, code reviews
- **Backend Engineers (2)**: Core services, KuzuDB integration
- **Frontend Engineer**: TUI development
- **DevOps Engineer**: CI/CD, deployment
- **QA Engineer**: Testing strategy, automation


### 9.2 Communication Plan


- Daily standups
- Weekly architecture reviews
- Bi-weekly demos
- Monthly retrospectives


## 10. Next Steps


1. **Week 1**: Team formation and environment setup
2. **Week 2**: Begin Phase 1 implementation
3. **Week 3**: First working prototype
4. **Week 4**: Initial user feedback session


This implementation guide provides a comprehensive blueprint for building a production-ready memory graph system. The combination of Domain Driven Design principles, modern Go practices, and the unique capabilities of KuzuDB creates a powerful foundation for next-generation AI applications.