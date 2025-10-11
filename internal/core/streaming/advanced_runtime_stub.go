//go:build !enterprise

package streaming

import (
	"context"

	streamingTypes "local/costscope/cmd/modules/streaming/types"
	"local/costscope/internal/core/enterprise"
	"local/costscope/internal/core/logging"
	persistence "local/costscope/internal/core/persistence"
	providers "local/costscope/internal/providers"
)

// Intentional enterprise stub for advanced streaming runtime components (job manager,
// persistent job manager, pipeline). Slim builds expose constructors that return
// disabled errors so call sites remain stable. Real implementations are behind the
// 'enterprise' build tag.

const FeatureAdvancedStreamingRuntime = "advanced_streaming_runtime"

func disabledAdvancedStreaming() error {
	enterprise.ObserveInvocation(FeatureAdvancedStreamingRuntime, false)
	enterprise.ObserveError(FeatureAdvancedStreamingRuntime, "disabled")
	return enterprise.DisabledError(FeatureAdvancedStreamingRuntime)
}

// Stub types (minimal) so signatures compile.
type DefaultJobManager struct{}

type PersistentJobManager struct{}

type Pipeline struct{}

type PipelineOptions struct{}

type Consumer func(ctx context.Context, msg *Message) error

type Message struct{}

// Constructors returning disabled errors ------------------------------------------------

func NewJobManager(_ *providers.ProviderManager, _ string) *DefaultJobManager { // providers imported in real build
	return &DefaultJobManager{}
}

func NewPersistentJobManager(_ persistence.Repository, _ *providers.ProviderManager, _ string) *PersistentJobManager {
	return &PersistentJobManager{}
}

func NewPipeline(_ *logging.Logger, _ PipelineOptions, _ Consumer) *Pipeline { return &Pipeline{} }

// Disabled behavior methods (subset used in tests / potential call sites) -------------

// Provide no-op methods so code linking against interfaces compiles; all return disabled error when invoked.
func (jm *DefaultJobManager) StartJob(_ *streamingTypes.StreamingJobConfig) (*streamingTypes.StreamingJobInfo, error) {
	return nil, disabledAdvancedStreaming()
}
func (jm *DefaultJobManager) PauseJob(string) error  { return disabledAdvancedStreaming() }
func (jm *DefaultJobManager) ResumeJob(string) error { return disabledAdvancedStreaming() }
func (jm *DefaultJobManager) StopJob(string) error   { return disabledAdvancedStreaming() }
func (jm *DefaultJobManager) GetJobStatus(string) (*streamingTypes.StreamingJobStatus, error) {
	return nil, disabledAdvancedStreaming()
}
func (jm *DefaultJobManager) ListJobs() ([]*streamingTypes.StreamingJobInfo, error) {
	return nil, disabledAdvancedStreaming()
}
func (jm *DefaultJobManager) Shutdown() error { return nil }

// Persistent manager parallels
func (pm *PersistentJobManager) StartJobPersistent(_ *streamingTypes.StreamingJobConfig) (*streamingTypes.StreamingJobInfo, error) {
	return nil, disabledAdvancedStreaming()
}

// Pipeline stubs (common minimal API)
func (p *Pipeline) Start() error                            { return disabledAdvancedStreaming() }
func (p *Pipeline) Publish(context.Context, *Message) error { return disabledAdvancedStreaming() }
func (p *Pipeline) Shutdown(context.Context) error          { return nil }
func (p *Pipeline) DeadLetters() []*Message                 { return nil }
func (p *Pipeline) SnapshotMetrics() Metrics                { return Metrics{} }

// Placeholder Metrics struct (real one behind enterprise build)
type Metrics struct{}
