package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	SharedServiceKind       = "SharedService"
	SharedServiceSliceKind  = "SharedServiceSlice"
	SharedServicePlanKind   = "SharedServicePlan"
	SharedServiceActionKind = "SharedServiceAction"
	SharedServiceFinalizer  = "finalizer.org.aerogear.sharedService"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceSpec   `json:"spec"`
	Status            SharedServiceStatus `json:"status"`
}

type SharedServiceStatus struct {
	CommonStatus
}

type CommonStatus struct {
	Phase SharedServiceStatusPhase `json:"phase,omitempty"`
	Ready bool                     `json:"ready"`
}

type SharedServiceSpec struct {
	MaxInstances      int    `json:"maxInstances"`
	MinInstances      int    `json:"minInstances"`
	SlicesPerInstance int    `json:"slicesPerInstance"`
	RequiredInstances int    `json:"requiredInstances"`
	ServiceType       string `json:"serviceType"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedService `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceSliceSpec   `json:"spec"`
	Status            SharedServiceSliceStatus `json:"status"`
}

type SharedServiceSliceSpec struct {
	ProvidedParams *runtime.RawExtension `json:"providedParams"`
	ServiceType    string                `json:"serviceType"`
	SliceNamespace string                `json:"sliceNamespace"`
}

type SharedServiceSliceStatus struct {
	CommonStatus
	// the ServiceInstanceID that represents the slice
	SliceServiceInstance string `json:"sliceServiceInstance"`
	// the ServiceInstanceID that represents the parent shared service
	SharedServiceInstance string `json:"sharedServiceInstance"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServiceSlice `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServicePlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServicePlanSpec   `json:"spec"`
	Status            SharedServicePlanStatus `json:"status"`
}

type SharedServicePlanSpec struct {
	ServiceType     string                      `json:"serviceType"`
	Name            string                      `json:"name"`
	ID              string                      `json:"id"`
	Description     string                      `json:"description"`
	Free            bool                        `json:"free"`
	BindParams      SharedServicePlanSpecParams `json:"bindParams"`
	ProvisionParams SharedServicePlanSpecParams `json:"provisionParams"`
}

type SharedServicePlanSpecParams struct {
	Schema     string                                         `json:"$schema"`
	Type       string                                         `json:"type"`
	Properties map[string]SharedServicePlanSpecParamsProperty `json:"properties"`
}

type SharedServicePlanSpecParamsProperty struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

type SharedServicePlanStatus struct {
	CommonStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServicePlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServicePlan `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceActionSpec   `json:"spec"`
	Status            SharedServiceActionStatus `json:"status"`
}

type SharedServiceActionSpec struct {
	ProvidedParams *runtime.RawExtension `json:"providedParams"`
	ServiceType    string                `json:"serviceType"`
}

type SharedServiceActionStatus struct {
	CommonStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServiceAction `json:"items"`
}

// StatusSharedConfig manages the capacity of a shared service
type StatusSharedConfig struct {
	SlicesPerInstance int `json:"slicesPerInstance"`
	CurrentSlices     int `json:"currentSlices"`
}

type SharedServiceStatusPhase string

const (
	SSPhaseNone                                  = ""
	SSPhaseAccepted     SharedServiceStatusPhase = "accepted"
	SSPhaseComplete     SharedServiceStatusPhase = "complete"
	SSPhaseProvisioning SharedServiceStatusPhase = "provisioning"
	SSPhaseProvisioned  SharedServiceStatusPhase = "provisioned"
)
