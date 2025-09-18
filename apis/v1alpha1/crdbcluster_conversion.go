package v1alpha1

import (
	"fmt"
	"github.com/cockroachdb/cockroach-operator/apis/v1beta1"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	BetaCrdbClusterNodeCloudProviderAnnotationKey = "crdb.cockroachlabs.com/cloudProvider"
	BetaCrdbClusterNodeRegionCodeAnnotationKey    = "crdb.cockroachlabs.com/regionCode"
)

var (
	defaultRegionCode    = "us-east1"
	defaultCloudProvider = "gcp"
)

// ConvertTo converts this v1alpha1 CrdbCluster to the Hub version (v1beta1)
func (src *CrdbCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.CrdbCluster)

	// Convert metadata
	dst.ObjectMeta = src.ObjectMeta

	// region code is not set for v1alpha1 clusters, so we are going to take it from annoations
	// if defined.
	if src.Annotations != nil {
		if region, ok := src.Annotations[BetaCrdbClusterNodeRegionCodeAnnotationKey]; ok {
			defaultRegionCode = region
		}
		if cloudProvider, ok := src.Annotations[BetaCrdbClusterNodeCloudProviderAnnotationKey]; ok {
			defaultCloudProvider = cloudProvider
		}
	}

	// Convert mode
	mode := v1beta1.MutableOnly
	rollingRestartDelay := metav1.Duration{Duration: 1 * time.Minute}
	// Convert spec
	dst.Spec = v1beta1.CrdbClusterSpec{
		Mode:                &mode,
		RollingRestartDelay: &rollingRestartDelay,
		TLSEnabled:          src.Spec.TLSEnabled,
		Template: v1beta1.CrdbNodeTemplate{
			Spec: v1beta1.CrdbNodeSpec{
				ServiceAccountName: "cockroach-operator-sa",
				Certificates: v1beta1.Certificates{
					ExternalCertificates: &v1beta1.ExternalCertificates{
						NodeSecretName:          fmt.Sprintf("%s-node", src.Name),
						RootSQLClientSecretName: fmt.Sprintf("%s-root", src.Name),
						HTTPSecretName:          fmt.Sprintf("%s-root", src.Name),
						CAConfigMapName:         "crdb-cockroachdb-ca-secret-crt",
					},
				},
			},
		},
	}

	// Convert nodes to regions
	if src.Spec.Nodes != 0 {
		dst.Spec.Regions = []v1beta1.CrdbClusterRegion{
			{
				Code:          defaultRegionCode, // Default region code for v1alpha1 clusters
				Nodes:         src.Spec.Nodes,
				CloudProvider: defaultCloudProvider,
				Namespace:     src.Namespace,
			},
		}
	}

	// Initialize flags map
	if dst.Spec.Template.Spec.Flags == nil {
		dst.Spec.Template.Spec.Flags = make(map[string]string)
	}

	// Convert cache and memory settings
	if src.Spec.Cache != "" {
		dst.Spec.Template.Spec.Flags["--cache"] = src.Spec.Cache
	}
	if src.Spec.MaxSQLMemory != "" {
		dst.Spec.Template.Spec.Flags["--max-sql-memory"] = src.Spec.MaxSQLMemory
	}

	// Convert additional args
	if src.Spec.AdditionalArgs != nil {
		for _, arg := range src.Spec.AdditionalArgs {
			// Simple conversion - in real implementation you might want to parse flags properly
			dst.Spec.Template.Spec.Flags[arg] = ""
		}
	}

	// Convert image
	if src.Spec.Image != nil {
		dst.Spec.Template.Spec.Image = src.Spec.Image.Name
	}

	// Convert resources
	dst.Spec.Template.Spec.ResourceRequirements = src.Spec.Resources

	// Convert pod settings
	dst.Spec.Template.Spec.PodLabels = src.Spec.AdditionalLabels
	dst.Spec.Template.Spec.PodAnnotations = src.Spec.AdditionalAnnotations
	dst.Spec.Template.Spec.Env = src.Spec.PodEnvVariables
	dst.Spec.Template.Spec.Affinity = src.Spec.Affinity
	dst.Spec.Template.Spec.TopologySpreadConstraints = src.Spec.TopologySpreadConstraints
	dst.Spec.Template.Spec.Tolerations = src.Spec.Tolerations
	dst.Spec.Template.Spec.NodeSelector = src.Spec.NodeSelector

	// Convert ports
	if src.Spec.GRPCPort != nil {
		dst.Spec.Template.Spec.GRPCPort = src.Spec.GRPCPort
	}
	if src.Spec.HTTPPort != nil {
		dst.Spec.Template.Spec.HTTPPort = src.Spec.HTTPPort
	}
	if src.Spec.SQLPort != nil {
		dst.Spec.Template.Spec.SQLPort = src.Spec.SQLPort
	}

	// Convert data store
	if !reflect.DeepEqual(src.Spec.DataStore, Volume{}) {
		dst.Spec.Template.Spec.DataStore = v1beta1.DataStore{}
		if src.Spec.DataStore.HostPath != nil {
			dst.Spec.Template.Spec.DataStore.VolumeSource = &corev1.VolumeSource{
				HostPath: src.Spec.DataStore.HostPath,
			}
		}
		if src.Spec.DataStore.VolumeClaim != nil {
			dst.Spec.Template.Spec.DataStore.VolumeClaimTemplate = &corev1.PersistentVolumeClaim{
				Spec: src.Spec.DataStore.VolumeClaim.PersistentVolumeClaimSpec,
			}
		}
	}

	// Convert status
	dst.Status = v1beta1.CrdbClusterStatus{
		Version:  src.Status.Version,
		Image:    src.Status.CrdbContainerImage,
		Region:   defaultRegionCode,
		Provider: defaultCloudProvider,
	}

	// Convert conditions
	if len(src.Status.Conditions) > 0 {
		dst.Status.Conditions = make([]v1beta1.ClusterCondition, len(src.Status.Conditions))
		for i, condition := range src.Status.Conditions {
			dst.Status.Conditions[i] = v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionType(condition.Type),
				Status:             condition.Status,
				LastTransitionTime: condition.LastTransitionTime,
			}
		}
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this v1alpha1 CrdbCluster
func (dst *CrdbCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.CrdbCluster)

	// Convert metadata
	dst.ObjectMeta = src.ObjectMeta

	// Convert spec
	dst.Spec = CrdbClusterSpec{
		TLSEnabled: src.Spec.TLSEnabled,
	}

	// Convert nodes from regions
	if len(src.Spec.Regions) > 0 {
		dst.Spec.Nodes = src.Spec.Regions[0].Nodes
	} else {
		dst.Spec.Nodes = int32(src.Spec.Nodes())
	}

	// Convert flags back to individual fields
	if src.Spec.Template.Spec.Flags != nil {
		if cache, ok := src.Spec.Template.Spec.Flags["--cache"]; ok {
			dst.Spec.Cache = cache
		}
		if maxSQL, ok := src.Spec.Template.Spec.Flags["--max-sql-memory"]; ok {
			dst.Spec.MaxSQLMemory = maxSQL
		}
	}

	// Convert image
	if src.Spec.Template.Spec.Image != "" {
		dst.Spec.Image = &PodImage{
			Name: src.Spec.Template.Spec.Image,
		}
	}

	// Convert resources
	dst.Spec.Resources = src.Spec.Template.Spec.ResourceRequirements

	// Convert pod settings
	dst.Spec.AdditionalLabels = src.Spec.Template.Spec.PodLabels
	dst.Spec.AdditionalAnnotations = src.Spec.Template.Spec.PodAnnotations
	dst.Spec.PodEnvVariables = src.Spec.Template.Spec.Env
	dst.Spec.Affinity = src.Spec.Template.Spec.Affinity
	dst.Spec.TopologySpreadConstraints = src.Spec.Template.Spec.TopologySpreadConstraints
	dst.Spec.Tolerations = src.Spec.Template.Spec.Tolerations
	dst.Spec.NodeSelector = src.Spec.Template.Spec.NodeSelector

	// Convert status
	dst.Status = CrdbClusterStatus{
		Version:            src.Status.Version,
		CrdbContainerImage: src.Status.Image,
	}

	// Convert conditions
	if len(src.Status.Conditions) > 0 {
		dst.Status.Conditions = make([]ClusterCondition, len(src.Status.Conditions))
		for i, condition := range src.Status.Conditions {
			dst.Status.Conditions[i] = ClusterCondition{
				Type:               ClusterConditionType(condition.Type),
				Status:             condition.Status,
				LastTransitionTime: condition.LastTransitionTime,
			}
		}
	}

	return nil
}
