/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resource

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/clusterstatus"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/gosimple/slug"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	RELATED_IMAGE_PREFIX         = "RELATED_IMAGE_COCKROACH_"
	NotSupportedVersion          = "not_supported_version"
	CrdbContainerImageAnnotation = "crdb.io/containerimage"
	CrdbVersionAnnotation        = "crdb.io/version"
	CrdbHistoryAnnotation        = "crdb.io/history"
	CrdbRestartAnnotation        = "crdb.io/restart"
	CrdbCertExpirationAnnotation = "crdb.io/certexpiration"
	CrdbRestartTypeAnnotation    = "crdb.io/restarttype"

	VersionCheckJobName = "vcheck"
)

func NewCluster(original *api.CrdbCluster) Cluster {
	cr := original.DeepCopy()
	cr.Default()

	timeNow := metav1.Now()
	condition.InitConditionsIfNeeded(&cr.Status, timeNow)
	clusterstatus.InitOperatorActionsIfNeeded(&cr.Status, timeNow)
	return Cluster{
		cr:       cr,
		initTime: timeNow,
	}
}

func ClusterPlaceholder(name string) *api.CrdbCluster {
	return &api.CrdbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

type Cluster struct {
	Fetcher

	cr       *api.CrdbCluster
	initTime metav1.Time
}

func (cluster Cluster) Unwrap() *api.CrdbCluster {
	return cluster.cr.DeepCopy()
}

func (cluster Cluster) SetTrue(ctype api.ClusterConditionType) {
	condition.SetTrue(ctype, &cluster.cr.Status, cluster.InitTime())
}

// True checks if the api.ClusterConditionType is true
func (cluster Cluster) True(ctype api.ClusterConditionType) bool {
	return condition.True(ctype, cluster.cr.Status.Conditions)
}

func (cluster Cluster) SetClusterStatusOnFirstReconcile() {
	clusterstatus.SetClusterStatusOnFirstReconcile(&cluster.cr.Status)
}
func (cluster Cluster) SetClusterStatus() {
	clusterstatus.SetClusterStatus(&cluster.cr.Status)
}
func (cluster Cluster) SetClusterVersion(version string) {
	cluster.cr.Status.Version = version
}
func (cluster Cluster) SetCrdbContainerImage(containerimage string) {
	cluster.cr.Status.CrdbContainerImage = containerimage
}
func (cluster Cluster) SetActionFailed(atype api.ActionType, errMsg string) {
	clusterstatus.SetActionFailed(atype, errMsg, &cluster.cr.Status)
}
func (cluster Cluster) ResetActionType(atype api.ActionType) {
	clusterstatus.ResetActionType(atype, &cluster.cr.Status)
}
func (cluster Cluster) SetActionFinished(atype api.ActionType) {
	clusterstatus.SetActionFinished(atype, &cluster.cr.Status)
}
func (cluster Cluster) SetActionUnknown(atype api.ActionType) {
	clusterstatus.SetActionUnknown(atype, &cluster.cr.Status)
}

func (cluster Cluster) UpdateDirectorState(state, activeActor string) (metav1.Time, int) {
	// Round to the nearest second; more granular information does not persist through various transformations
	now := metav1.NewTime(time.Now().Round(time.Second))

	cluster.cr.Status.DirectorState = state
	cluster.cr.Status.DirectorStateUpdatedAt = now
	cluster.cr.Status.DirectorObservedGeneration++
	cluster.cr.Status.ActiveActor = activeActor

	return now, cluster.cr.Status.DirectorObservedGeneration
}

func (cluster Cluster) Failed(atype api.ActionType) bool {
	return clusterstatus.Failed(atype, cluster.Status().OperatorActions)
}
func (cluster Cluster) SetFalse(ctype api.ClusterConditionType) {
	condition.SetFalse(ctype, &cluster.cr.Status, cluster.InitTime())
}

func (cluster Cluster) Spec() *api.CrdbClusterSpec {
	return cluster.cr.Spec.DeepCopy()
}

func (cluster Cluster) Status() *api.CrdbClusterStatus {
	return cluster.cr.Status.DeepCopy()
}

func (cluster Cluster) Name() string {
	return cluster.cr.Name
}

func (cluster Cluster) Namespace() string {
	return cluster.cr.Namespace
}

func (cluster Cluster) ObjectKey() types.NamespacedName {
	return types.NamespacedName{Namespace: cluster.Namespace(), Name: cluster.Name()}
}

func (cluster Cluster) InitTime() metav1.Time {
	return cluster.initTime
}

func (cluster Cluster) DiscoveryServiceName() string {
	return cluster.Name()
}

func (cluster Cluster) PublicServiceName() string {
	return fmt.Sprintf("%s-public", cluster.Name())
}

func (cluster Cluster) StatefulSetName() string {
	return cluster.Name()
}

func (cluster Cluster) JobName() string {
	slug.MaxLength = 63
	return slug.Make(fmt.Sprintf("%s-%s-%d", cluster.Name(), VersionCheckJobName, getTimeHashInMinutes(time.Now())))
}
func getTimeHashInMinutes(scheduledTime time.Time) int64 {
	return scheduledTime.Unix() / 60
}
func (cluster Cluster) IsSupportedImage() bool {
	image := cluster.GetCockroachDBImageName()
	return !strings.EqualFold(image, NotSupportedVersion)
}
func (cluster Cluster) LookupSupportedVersion(version string) (string, bool) {
	if version == "" {
		return "", false
	}
	supportedVersions := getSupportedCrdbVersions()
	for _, v := range supportedVersions {
		if strings.EqualFold(version, v) {
			return v, true
		}
	}
	return "", false
}

//GetVersionAnnotation  gets the current version of the cluster  retrieved by version checker action
func (cluster Cluster) GetVersionAnnotation() string {
	return cluster.getAnnotation(CrdbVersionAnnotation)
}

func (cluster Cluster) GetAnnotationContainerImage() string {
	return cluster.getAnnotation(CrdbContainerImageAnnotation)
}

func (cluster Cluster) GetAnnotationRestartType() string {
	return cluster.getAnnotation(CrdbRestartTypeAnnotation)
}

func (cluster Cluster) GetAnnotationHistory() string {
	return cluster.getAnnotation(CrdbHistoryAnnotation)
}

func (cluster Cluster) getAnnotation(key string) string {
	if val, ok := cluster.cr.Annotations[key]; !ok {
		return ""
	} else {
		return val
	}
}
func (cluster Cluster) SetAnnotationVersion(version string) {
	if cluster.cr.Annotations == nil {
		cluster.cr.Annotations = make(map[string]string)
	}
	cluster.cr.Annotations[CrdbVersionAnnotation] = version
}
func (cluster Cluster) SetAnnotationContainerImage(containerimage string) {
	if cluster.cr.Annotations == nil {
		cluster.cr.Annotations = make(map[string]string)
	}
	cluster.cr.Annotations[CrdbContainerImageAnnotation] = containerimage
}
func (cluster Cluster) SetAnnotationCertExpiration(certExpiration string) {
	if cluster.cr.Annotations == nil {
		cluster.cr.Annotations = make(map[string]string)
	}
	cluster.cr.Annotations[CrdbCertExpirationAnnotation] = certExpiration
}
func (cluster Cluster) DeleteRestartTypeAnnotation() {
	if cluster.cr.Annotations == nil {
		return
	}
	delete(cluster.cr.Annotations, CrdbRestartTypeAnnotation)
}

func (cluster Cluster) GetCockroachDBImageName() string {
	supportedImages := getSupportedCrdbImages()
	if cluster.Spec().CockroachDBVersion != "" {
		if version, ok := cluster.LookupSupportedVersion(cluster.Spec().CockroachDBVersion); ok {
			if version == "" {
				return NotSupportedVersion
			}
			if image, ok := supportedImages[version]; ok {
				if image == "" {
					return NotSupportedVersion
				}
				return image
			}
		}
		return NotSupportedVersion
	}
	//we validate the version after the job runs with exec
	return cluster.Spec().Image.Name
}

func (cluster Cluster) NodeTLSSecretName() string {
	return fmt.Sprintf("%s-node", cluster.Name())
}

func (cluster Cluster) ClientTLSSecretName() string {
	return fmt.Sprintf("%s-root", cluster.Name())
}
func (cluster Cluster) CASecretName() string {
	return fmt.Sprintf("%s-ca", cluster.Name())
}

func (cluster Cluster) Domain() string {
	return "svc.cluster.local"
}

func (cluster Cluster) SecureMode() string {
	if cluster.Spec().TLSEnabled {
		return "--certs-dir=/cockroach/cockroach-certs/"
	}

	return "--insecure"
}

func (cluster Cluster) SpecChanged(refreshedCluster *api.CrdbCluster) (bool, error) {
	return !reflect.DeepEqual(cluster.cr.Spec, refreshedCluster.Spec), nil
}

// TODO add error handling to ensure that env variables are set correctly and
// that we have a min number of them

// getSupportedCrdbImages will dynamic build an slice using the env var added in the operator.yaml
// We will add all the env var that start with RELATED_IMAGE
func getSupportedCrdbImages() map[string]string {
	crdbSupportedImages := make(map[string]string)
	supportedVersions := getSupportedCrdbVersions()
	for _, v := range supportedVersions {
		envVar := fmt.Sprintf("%s%s", RELATED_IMAGE_PREFIX, strings.ReplaceAll(v, ".", "_"))
		crdbSupportedImages[v] = os.Getenv(envVar)
	}
	return crdbSupportedImages
}
func getSupportedCrdbVersions() []string {
	supportedVersions := make([]string, 0)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], RELATED_IMAGE_PREFIX) {
			version := strings.ReplaceAll(pair[0], RELATED_IMAGE_PREFIX, "")
			version = strings.ReplaceAll(version, "_", ".")
			supportedVersions = append(supportedVersions, version)
		}
	}
	return supportedVersions
}
