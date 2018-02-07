package pipeline

import (
	"fmt"
	"sync"

	"github.com/gaia-pipeline/gaia"
)

// BuildPipeline is the interface for pipelines which
// are not yet compiled.
type BuildPipeline interface {
	// PrepareEnvironment prepares the environment before we start the
	// build process.
	PrepareEnvironment(*gaia.CreatePipeline) error

	// ExecuteBuild executes the compiler and tracks the status of
	// the compiling process.
	ExecuteBuild(*gaia.CreatePipeline) error

	// CopyBinary copies the result from the compile process
	// to the plugins folder.
	CopyBinary(*gaia.CreatePipeline) error
}

// ActivePipelines holds all active pipelines.
// ActivePipelines can be safely shared between goroutines.
type ActivePipelines struct {
	sync.RWMutex

	// All active pipelines
	Pipelines []gaia.Pipeline
}

const (
	// Temp folder where we store our temp files during build pipeline.
	tmpFolder = "tmp"

	// Max minutes until the build process will be interrupted and marked as failed
	maxTimeoutMinutes = 60

	// typeDelimiter defines the delimiter in the file name to define
	// the pipeline type.
	typeDelimiter = "_"
)

var (
	// GlobalActivePipelines holds globally all current active pipleines.
	GlobalActivePipelines *ActivePipelines
)

// NewBuildPipeline creates a new build pipeline for the given
// pipeline type.
func NewBuildPipeline(t gaia.PipelineType) BuildPipeline {
	var bP BuildPipeline

	// Create build pipeline for given pipeline type
	switch t {
	case gaia.GOLANG:
		bP = &BuildPipelineGolang{
			Type: t,
		}
	}

	return bP
}

// NewActivePipelines creates a new instance of ActivePipelines
func NewActivePipelines() *ActivePipelines {
	ap := &ActivePipelines{
		Pipelines: make([]gaia.Pipeline, 0),
	}

	return ap
}

// Append appends a new pipeline to ActivePipelines.
func (ap *ActivePipelines) Append(p gaia.Pipeline) {
	ap.Lock()
	defer ap.Unlock()

	ap.Pipelines = append(ap.Pipelines, p)
}

// Iter iterates over the pipelines in the concurrent slice.
func (ap *ActivePipelines) Iter() <-chan gaia.Pipeline {
	c := make(chan gaia.Pipeline)

	go func() {
		ap.Lock()
		defer ap.Unlock()
		for _, pipeline := range ap.Pipelines {
			c <- pipeline
		}
		close(c)
	}()

	return c
}

// Contains checks if the given pipeline name has been already appended
// to the given ActivePipelines instance.
func (ap *ActivePipelines) Contains(n string) bool {
	for pipeline := range ap.Iter() {
		if pipeline.Name == n {
			return true
		}
	}

	return false
}

// appendTypeToName appends the type to the output binary name.
// This allows us later to define the pipeline type by the name.
func appendTypeToName(n string, pType gaia.PipelineType) string {
	return fmt.Sprintf("%s%s%s", n, typeDelimiter, pType.String())
}
