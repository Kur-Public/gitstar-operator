package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitStarSpec defines the desired state of GitStar
type GitStarSpec struct {
    // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
    // Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
    // Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
    RepoDomain string `json:"repo_domain"`
    RepoName   string `json:"repo_name"`
}

// GitStarStatus defines the observed state of GitStar
type GitStarStatus struct {
    // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
    // Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
    // Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
    StarNumber int64 `json:"star_number"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitStar is the Schema for the gitstars API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gitstars,scope=Namespaced
type GitStar struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   GitStarSpec   `json:"spec,omitempty"`
    Status GitStarStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitStarList contains a list of GitStar
type GitStarList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []GitStar `json:"items"`
}

func init() {
    SchemeBuilder.Register(&GitStar{}, &GitStarList{})
}
