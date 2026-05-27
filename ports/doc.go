// Package ports defines integration contracts around the iteration engine.
//
// Implement these interfaces in business services, queues, workers, storage
// layers, or optional LLM adapters. The engine package depends on ports; ports
// do not depend on engine orchestration.
package ports
