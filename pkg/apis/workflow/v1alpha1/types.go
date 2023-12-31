package v1alpha1

import (
	"encoding/json"
	"fmt"
	"hash/fnv"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemplateType is the type of a template
type TemplateType string

// Possible template types
const (
	TemplateTypeContainer TemplateType = "Container"
	TemplateTypeSteps     TemplateType = "Steps"
	TemplateTypeScript    TemplateType = "Script"
	TemplateTypeResource  TemplateType = "Resource"
)

// NodePhase is a label for the condition of a node at the current time.
type NodePhase string

// Workflow and node statuses
const (
	NodeRunning   NodePhase = "Running"
	NodeSucceeded NodePhase = "Succeeded"
	NodeSkipped   NodePhase = "Skipped"
	NodeFailed    NodePhase = "Failed"
	NodeError     NodePhase = "Error"
)

// Workflow is the definition of a workflow resource
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec              WorkflowSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status            WorkflowStatus `json:"status" protobuf:"bytes,3,opt,name=status"`
}

// WorkflowList is list of Workflow resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Workflow `json:"items" protobuf:"bytes,2,opt,name=items"`
}

// WorkflowSpec is the specification of a Workflow.
type WorkflowSpec struct {
	// Templates is a list of workflow templates used in a workflow
	Templates []Template `json:"templates" protobuf:"bytes,1,opt,name=templates"`

	// Entrypoint is a template reference to the starting point of the workflow
	Entrypoint string `json:"entrypoint" protobuf:"bytes,2,opt,name=entrypoint"`

	// Arguments contain the parameters and artifacts sent to the workflow entrypoint
	// Parameters are referencable globally using the 'workflow' variable prefix.
	// e.g. {{workflow.parameters.myparam}}
	Arguments Arguments `json:"arguments,omitempty" protobuf:"bytes,3,opt,name=arguments"`

	// ServiceAccountName is the name of the ServiceAccount to run all pods of the workflow as.
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,4,opt,name=serviceAccountName"`

	// Volumes is a list of volumes that can be mounted by containers in a workflow.
	Volumes []apiv1.Volume `json:"volumes,omitempty" protobuf:"bytes,5,opt,name=volumes"`

	// VolumeClaimTemplates is a list of claims that containers are allowed to reference.
	// The Workflow controller will create the claims at the beginning of the workflow
	// and delete the claims upon completion of the workflow
	VolumeClaimTemplates []apiv1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,6,opt,name=volumeClaimTemplates"`

	// NodeSelector is a selector which will result in all pods of the workflow
	// to be scheduled on the selected node(s). This is able to be overridden by
	// a nodeSelector specified in the template.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,opt,name=nodeSelector"`

	// Affinity sets the scheduling constraints for all pods in the workflow.
	// Can be overridden by an affinity specified in the template
	Affinity *apiv1.Affinity `json:"affinity,omitempty" protobuf:"bytes,8,opt,name=affinity"`

	// ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images
	// in pods that reference this ServiceAccount. ImagePullSecrets are distinct from Secrets because Secrets
	// can be mounted in the pod, but ImagePullSecrets are only accessed by the kubelet.
	// More info: https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []apiv1.LocalObjectReference `json:"imagePullSecrets,omitempty" protobuf:"bytes,9,opt,name=imagePullSecrets"`

	// OnExit is a template reference which is invoked at the end of the
	// workflow, irrespective of the success, failure, or error of the
	// primary workflow.
	OnExit string `json:"onExit,omitempty" protobuf:"bytes,10,opt,name=onExit"`
}

// Template is a reusable and composable unit of execution in a workflow
type Template struct {
	// Name is the name of the template
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Inputs describe what inputs parameters and artifacts are supplied to this template
	Inputs Inputs `json:"inputs,omitempty" protobuf:"bytes,2,opt,name=inputs"`

	// Outputs describe the parameters and artifacts that this template produces
	Outputs Outputs `json:"outputs,omitempty" protobuf:"bytes,3,opt,name=outputs"`

	// NodeSelector is a selector to schedule this step of the workflow to be
	// run on the selected node(s). Overrides the selector set at the workflow level.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,4,opt,name=nodeSelector"`

	// Affinity sets the pod's scheduling constraints
	// Overrides the affinity set at the workflow level (if any)
	Affinity *apiv1.Affinity `json:"affinity,omitempty" protobuf:"bytes,5,opt,name=affinity"`

	// Deamon will allow a workflow to proceed to the next step so long as the container reaches readiness
	Daemon *bool `json:"daemon,omitempty" protobuf:"bytes,6,opt,name=daemon"`

	// Steps define a series of sequential/parallel workflow steps
	Steps [][]WorkflowStep `json:"steps,omitempty" protobuf:"bytes,7,opt,name=steps"`

	// Container is the main container image to run in the pod
	Container *apiv1.Container `json:"container,omitempty" protobuf:"bytes,8,opt,name=container"`

	// Script runs a portion of code against an interpreter
	Script *Script `json:"script,omitempty" protobuf:"bytes,9,opt,name=script"`

	// Resource template subtype which can run k8s resources
	Resource *ResourceTemplate `json:"resource,omitempty" protobuf:"bytes,10,opt,name=resource"`

	// Sidecars is a list of containers which run alongside the main container
	// Sidecars are automatically killed when the main container completes
	Sidecars []Sidecar `json:"sidecars,omitempty" protobuf:"bytes,11,opt,name=sidecars"`

	// Location in which all files related to the step will be stored (logs, artifacts, etc...).
	// Can be overridden by individual items in Outputs. If omitted, will use the default
	// artifact repository location configured in the controller, appended with the
	// <workflowname>/<nodename> in the key.
	ArchiveLocation *ArtifactLocation `json:"archiveLocation,omitempty" protobuf:"bytes,12,opt,name=archieveLocation"`

	// Optional duration in seconds relative to the StartTime that the pod may be active on a node
	// before the system actively tries to terminate the pod; value must be positive integer
	// This field is only applicable to container and script templates.
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"bytes,13,opt,name=activeDeadlineSeconds"`

	// RetryStrategy describes how to retry a template when it fails
	RetryStrategy *RetryStrategy `json:"retryStrategy,omitempty" protobuf:"bytes,14,opt,name=retryStrategy"`
}

// Inputs are the mechanism for passing parameters, artifacts, volumes from one template to another
type Inputs struct {
	// Parameters are a list of parameters passed as inputs
	Parameters []Parameter `json:"parameters,omitempty" protobuf:"bytes,1,opt,name=parameters"`

	// Artifact are a list of artifacts passed as inputs
	Artifacts []Artifact `json:"artifacts,omitempty" protobuf:"bytes,2,opt,name=artifacts"`
}

// Parameter indicate a passed string parameter to a service template with an optional default value
type Parameter struct {
	// Name is the parameter name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Default is the default value to use for an input parameter if a value was not supplied
	Default *string `json:"default,omitempty" protobuf:"bytes,2,opt,name=default"`

	// Value is the literal value to use for the parameter.
	// If specified in the context of an input parameter, the value takes precedence over any passed values
	Value *string `json:"value,omitempty" protobuf:"bytes,3,opt,name=value"`

	// ValueFrom is the source for the output parameter's value
	ValueFrom *ValueFrom `json:"valueFrom,omitempty" protobuf:"bytes,4,opt,name=valueFrom"`
}

// ValueFrom describes a location in which to obtain the value to a parameter
type ValueFrom struct {
	// Path in the container to retrieve an output parameter value from in container templates
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`

	// JSONPath of a resource to retrieve an output parameter value from in resource templates
	JSONPath string `json:"jsonPath,omitempty" protobuf:"bytes,2,opt,name=jsonPath"`

	// JQFilter expression against the resource object in resource templates
	JQFilter string `json:"jqFilter,omitempty" protobuf:"bytes,3,opt,name=jqFilter"`

	// Parameter reference to a step or dag task in which to retrieve an output parameter value from
	// (e.g. '{{steps.mystep.outputs.myparam}}')
	Parameter string `json:"parameter,omitempty" protobuf:"bytes,4,opt,name=parameter"`
}

// Artifact indicates an artifact to place at a specified path
type Artifact struct {
	// name of the artifact. must be unique within a template's inputs/outputs.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Path is the container path to the artifact
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// mode bits to use on this file, must be a value between 0 and 0777
	// set when loading input artifacts.
	Mode *int32 `json:"mode,omitempty" protobuf:"bytes,3,opt,name=mode"`

	// From allows an artifact to reference an artifact from a previous step
	From string `json:"from,omitempty" protobuf:"bytes,4,opt,name=from"`

	// ArtifactLocation contains the location of the artifact
	ArtifactLocation `json:",inline" protobuf:"bytes,5,opt,name=artifactLocation"`
}

// ArtifactLocation describes a location for a single or multiple artifacts.
// It is used as single artifact in the context of inputs/outputs (e.g. outputs.artifacts.artname).
// It is also used to describe the location of multiple artifacts such as the archive location
// of a single workflow step, which the executor will use as a default location to store its files.
type ArtifactLocation struct {
	// S3 contains S3 artifact location details
	S3 *S3Artifact `json:"s3,omitempty" protobuf:"bytes,1,opt,name=s3"`

	// Git contains git artifact location details
	Git *GitArtifact `json:"git,omitempty" protobuf:"bytes,2,opt,name=git"`

	// HTTP contains HTTP artifact location details
	HTTP *HTTPArtifact `json:"http,omitempty" protobuf:"bytes,3,opt,name=http"`

	// Artifactory contains artifactory artifact location details
	Artifactory *ArtifactoryArtifact `json:"artifactory,omitempty" protobuf:"bytes,4,opt,name=artifactory"`

	// Raw contains raw artifact location details
	Raw *RawArtifact `json:"raw,omitempty" protobuf:"bytes,5,opt,name=raw"`
}

// Outputs hold parameters, artifacts, and results from a step
type Outputs struct {
	// Parameters holds the list of output parameters produced by a step
	Parameters []Parameter `json:"parameters,omitempty" protobuf:"bytes,1,opt,name=parameters"`

	// Artifacts holds the list of output artifacts produced by a step
	Artifacts []Artifact `json:"artifacts,omitempty" protobuf:"bytes,2,opt,name=artifacts"`

	// Result holds the result (stdout) of a script template
	Result *string `json:"result,omitempty" protobuf:"bytes,3,opt,name=result"`
}

// WorkflowStep is a reference to a template to execute in a series of step
type WorkflowStep struct {
	// Name of the step
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Template is a reference to the template to execute as the step
	Template string `json:"template,omitempty" protobuf:"bytes,2,opt,name=template"`

	// Arguments hold arguments to the template
	Arguments Arguments `json:"arguments,omitempty" protobuf:"bytes,3,opt,name=arguments"`

	// WithItems expands a step into multiple parallel steps from the items in the list
	// TODO(jessesuen): kube-openapi cannot handle interfaces{}
	// The right solution is to create a new MapOrString struct like IntOrString
	// See: k8s.io/apimachinery/pkg/util/intstr/intstr.go
	// +k8s:openapi-gen=false
	WithItems []Item `json:"withItems,omitempty" protobuf:"bytes,4,opt,name=withItems"`

	// WithParam expands a step into from the value in the parameter
	WithParam string `json:"withParam,omitempty" protobuf:"bytes,5,opt,name=withParam"`

	// When is an expression in which the step should conditionally execute
	When string `json:"when,omitempty" protobuf:"bytes,6,opt,name=when"`
}

// Arguments to a template
type Arguments struct {
	// Parameters is the list of parameters to pass to the template or workflow
	Parameters []Parameter `json:"parameters,omitempty" protobuf:"bytes,1,opt,name=parameters"`

	// Artifacts is the list of artifacts to pass to the template or workflow
	Artifacts []Artifact `json:"artifacts,omitempty" protobuf:"bytes,2,opt,name=artifacts"`
}

// Sidecar is a container which runs alongside the main container
type Sidecar struct {
	apiv1.Container `json:",inline" protobuf:"bytes,1,opt,name=container"`

	SidecarOptions `json:",inline" protobuf:"bytes,2,opt,name=sizecarOptions"`
}

// SidecarOptions provide a way to customize the behavior of a sidecar and how it
// affects the main container.
type SidecarOptions struct {

	// MirrorVolumeMounts will mount the same volumes specified in the main container
	// to the sidecar (including artifacts), at the same mountPaths. This enables
	// dind daemon to partially see the same filesystem as the main container in
	// order to use features such as docker volume binding
	MirrorVolumeMounts *bool `json:"mirrorVolumeMounts,omitempty" protobuf:"bytes,1,opt,name=mirrorVolumeMounts"`

	// Other sidecar options to consider:
	// * Lifespan - allow a sidecar to live longer than the main container and run to completion.
	// * PropagateFailure - if a sidecar fails, also fail the step
}

// WorkflowStatus contains overall status information about a workflow
// +k8s:openapi-gen=false
type WorkflowStatus struct {
	// Phase a simple, high-level summary of where the workflow is in its lifecycle.
	Phase NodePhase `json:"phase" protobuf:"bytes,1,opt,name=phase"`

	// Time at which this workflow started
	StartedAt metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,2,opt,name=startedAt"`

	// Time at which this workflow completed
	FinishedAt metav1.Time `json:"finishedAt,omitempty" protobuf:"bytes,3,opt,name=finishedAt"`

	// A human readable message indicating details about why the workflow is in this condition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`

	// Nodes is a mapping between a node ID and the node's status.
	Nodes map[string]NodeStatus `json:"nodes" protobuf:"bytes,5,opt,name=nodes"`

	// PersistentVolumeClaims tracks all PVCs that were created as part of the workflow.
	// The contents of this list are drained at the end of the workflow.
	PersistentVolumeClaims []apiv1.Volume `json:"persistentVolumeClaims,omitempty" protobuf:"bytes,6,opt,name=persistentVolumeClaims"`
}

// GetNodesWithRetries returns a list of nodes that have retries.
func (wfs *WorkflowStatus) GetNodesWithRetries() []NodeStatus {
	var nodesWithRetries []NodeStatus
	for _, node := range wfs.Nodes {
		if node.RetryStrategy != nil {
			nodesWithRetries = append(nodesWithRetries, node)
		}
	}
	return nodesWithRetries
}

// RetryStrategy provides controls on how to retry a workflow step
type RetryStrategy struct {
	// Limit is the maximum number of attempts when retrying a container
	Limit *int32 `json:"limit,omitempty" protobuf:"varint,1,opt,name=limit"`
}

// NodeStatus contains status information about an individual node in the workflow
// +k8s:openapi-gen=false
type NodeStatus struct {
	// ID is a unique identifier of a node within the worklow
	// It is implemented as a hash of the node name, which makes the ID deterministic
	ID string `json:"id" protobuf:"bytes,1,opt,name=id"`

	// Name is a human readable representation of the node in the node tree
	// It can represent a container, step group, or the entire workflow
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// Phase a simple, high-level summary of where the node is in its lifecycle.
	// Can be used as a state machine.
	Phase NodePhase `json:"phase" protobuf:"bytes,3,opt,name=phase"`

	// A human readable message indicating details about why the node is in this condition.
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`

	// Time at which this node started
	StartedAt metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,5,opt,name=startedAt"`

	// Time at which this node completed
	FinishedAt metav1.Time `json:"finishedAt,omitempty" protobuf:"bytes,6,opt,name=finishedAt"`

	// PodIP captures the IP of the pod for daemoned steps
	PodIP string `json:"podIP,omitempty" protobuf:"bytes,7,opt,name=podIP"`

	// Daemoned tracks whether or not this node was daemoned and need to be terminated
	Daemoned *bool `json:"daemoned,omitempty" protobuf:"bytes,8,opt,name=daemoned"`

	// RetryStrategy contains retry information about the node
	RetryStrategy *RetryStrategy `json:"retryStrategy,omitempty" protobuf:"bytes,9,opt,name=retryStrategy"`

	// Outputs captures output parameter values and artifact locations
	Outputs *Outputs `json:"outputs,omitempty" protobuf:"bytes,10,opt,name=outputs"`

	// Children is a list of child node IDs
	Children []string `json:"children,omitempty" protobuf:"bytes,11,opt,name=children"`
}

func (n NodeStatus) String() string {
	return fmt.Sprintf("%s (%s)", n.Name, n.ID)
}

// Completed returns whether or not the node has completed execution
func (n NodeStatus) Completed() bool {
	return n.Phase == NodeSucceeded ||
		n.Phase == NodeFailed ||
		n.Phase == NodeError ||
		n.Phase == NodeSkipped
}

// IsDaemoned returns whether or not the node is deamoned
func (n NodeStatus) IsDaemoned() bool {
	if n.Daemoned == nil || !*n.Daemoned {
		return false
	}
	return true
}

// Successful returns whether or not this node completed successfully
func (n NodeStatus) Successful() bool {
	return n.Phase == NodeSucceeded || n.Phase == NodeSkipped
}

// CanRetry returns whether the node should be retried or not.
func (n NodeStatus) CanRetry() bool {
	// TODO(shri): Check if there are some 'unretryable' errors.
	return n.Completed() && !n.Successful()
}

// S3Bucket contains the access information required for interfacing with an S3 bucket
type S3Bucket struct {
	// Endpoint is the hostname of the bucket endpoint
	Endpoint string `json:"endpoint" protobuf:"bytes,1,opt,name=endpoint"`

	// Bucket is the name of the bucket
	Bucket string `json:"bucket" protobuf:"bytes,2,opt,name=bucket"`

	// Region contains the optional bucket region
	Region string `json:"region,omitempty" protobuf:"bytes,3,opt,name=region"`

	// Insecure will connect to the service with TLS
	Insecure *bool `json:"insecure,omitempty" protobuf:"varint,4,opt,name=insecure"`

	// AccessKeySecret is the secret selector to the bucket's access key
	AccessKeySecret apiv1.SecretKeySelector `json:"accessKeySecret" protobuf:"bytes,5,opt,name=accessKeySecret"`

	// SecretKeySecret is the secret selector to the bucket's secret key
	SecretKeySecret apiv1.SecretKeySelector `json:"secretKeySecret" protobuf:"bytes,6,opt,name=secretKeySecret"`
}

// S3Artifact is the location of an S3 artifact
type S3Artifact struct {
	S3Bucket `json:",inline" protobuf:"bytes,1,opt,name=s3Bucket"`

	// Key is the key in the bucket where the artifact resides
	Key string `json:"key" protobuf:"bytes,2,opt,name=key"`
}

// GitArtifact is the location of an git artifact
type GitArtifact struct {
	// Repo is the git repository
	Repo string `json:"repo" protobuf:"bytes,1,opt,name=repo"`

	// Revision is the git commit, tag, branch to checkout
	Revision string `json:"revision,omitempty" protobuf:"bytes,2,opt,name=revision"`

	// UsernameSecret is the secret selector to the repository username
	UsernameSecret *apiv1.SecretKeySelector `json:"usernameSecret,omitempty" protobuf:"bytes,3,opt,name=usernameSecret"`

	// PasswordSecret is the secret selector to the repository password
	PasswordSecret *apiv1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,4,opt,name=passwordSecret"`
}

// ArtifactoryAuth describes the secret selectors required for authenticating to artifactory
type ArtifactoryAuth struct {
	// UsernameSecret is the secret selector to the repository username
	UsernameSecret *apiv1.SecretKeySelector `json:"usernameSecret,omitempty" protobuf:"bytes,1,opt,name=usernameSecret"`

	// PasswordSecret is the secret selector to the repository password
	PasswordSecret *apiv1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,2,opt,name=passwordSecret"`
}

// ArtifactoryArtifact is the location of an artifactory artifact
type ArtifactoryArtifact struct {
	// URL of the artifact
	URL             string `json:"url" protobuf:"bytes,1,opt,name=url"`
	ArtifactoryAuth `json:",inline" protobuf:"bytes,2,opt,name=artifactoryAuth"`
}

// RawArtifact allows raw string content to be placed as an artifact in a container
type RawArtifact struct {
	// Data is the string contents of the artifact
	Data string `json:"data" protobuf:"bytes,1,opt,name=data"`
}

// HTTPArtifact allows an file served on HTTP to be placed as an input artifact in a container
type HTTPArtifact struct {
	// URL of the artifact
	URL string `json:"url" protobuf:"bytes,1,opt,name=url"`
}

// Script is a template subtype to enable scripting through code steps
type Script struct {
	// Image is the container image to run
	Image string `json:"image" protobuf:"bytes,1,opt,name=image"`

	// Command is the interpreter coommand to run (e.g. [python])
	Command []string `json:"command" protobuf:"bytes,2,opt,name=command"`

	// Source contains the source code of the script to execute
	Source string `json:"source" protobuf:"bytes,3,opt,name=source"`
}

// ResourceTemplate is a template subtype to manipulate kubernetes resources
type ResourceTemplate struct {
	// Action is the action to perform to the resource.
	// Must be one of: get, create, apply, delete, replace
	Action string `json:"action" protobuf:"bytes,1,opt,name=action"`

	// Manifest contains the kubernetes manifest
	Manifest string `json:"manifest" protobuf:"bytes,2,opt,name=manifest"`

	// SuccessCondition is a label selector expression which describes the conditions
	// of the k8s resource in which it is acceptable to proceed to the following step
	SuccessCondition string `json:"successCondition,omitempty" protobuf:"bytes,3,opt,name=successCondition"`

	// FailureCondition is a label selector expression which describes the conditions
	// of the k8s resource in which the step was considered failed
	FailureCondition string `json:"failureCondition,omitempty" protobuf:"bytes,4,opt,name=failureCondition"`
}

// GetType returns the type of this template
func (tmpl *Template) GetType() TemplateType {
	if tmpl.Container != nil {
		return TemplateTypeContainer
	}
	if tmpl.Steps != nil {
		return TemplateTypeSteps
	}
	if tmpl.Script != nil {
		return TemplateTypeScript
	}
	if tmpl.Resource != nil {
		return TemplateTypeResource
	}
	return "Unknown"
}

// GetArtifactByName returns an input artifact by its name
func (in *Inputs) GetArtifactByName(name string) *Artifact {
	for _, art := range in.Artifacts {
		if art.Name == name {
			return &art
		}
	}
	return nil
}

// GetParameterByName returns an input parameter by its name
func (in *Inputs) GetParameterByName(name string) *Parameter {
	for _, param := range in.Parameters {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

// HasOutputs returns whether or not there are any outputs
func (out *Outputs) HasOutputs() bool {
	if out.Result != nil {
		return true
	}
	if len(out.Artifacts) > 0 {
		return true
	}
	if len(out.Parameters) > 0 {
		return true
	}
	return false
}

// GetArtifactByName retrieves an artifact by its name
func (args *Arguments) GetArtifactByName(name string) *Artifact {
	for _, art := range args.Artifacts {
		if art.Name == name {
			return &art
		}
	}
	return nil
}

// GetParameterByName retrieves a parameter by its name
func (args *Arguments) GetParameterByName(name string) *Parameter {
	for _, param := range args.Parameters {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

// HasLocation whether or not an artifact has a location defined
func (a *Artifact) HasLocation() bool {
	return a.S3 != nil || a.Git != nil || a.HTTP != nil || a.Artifactory != nil || a.Raw != nil
}

// DeepCopyInto is an custom deepcopy function to deal with our use of the interface{} type
func (in *WorkflowStep) DeepCopyInto(out *WorkflowStep) {
	inBytes, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(inBytes, out)
	if err != nil {
		panic(err)
	}
}

// GetTemplate retrieves a defined template by its name
func (wf *Workflow) GetTemplate(name string) *Template {
	for _, t := range wf.Spec.Templates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

// NodeID creates a deterministic node ID based on a node name
func (wf *Workflow) NodeID(name string) string {
	if name == wf.ObjectMeta.Name {
		return wf.ObjectMeta.Name
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	return fmt.Sprintf("%s-%v", wf.ObjectMeta.Name, h.Sum32())
}
