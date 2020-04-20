package resource

import (
	"context"
	"fmt"

	v1 "k8s.io/api/batch/v1"
	batchv1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "gitstar-operator/pkg/apis/app/v1"
)

const (
	ENVGitStarName      = "git_star_name"
	ENVGitStarNameSpace = "git_star_name_space"
)

var (
	CronJobHistoryLimit int32 = 5

	log = logf.Log.WithName("controller_gitstar")
)

// newCronJobForCR
func NewCronJobForCR(cr *appv1.GitStar) *batchv1.CronJob {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenerateCronJobName(cr),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			// TODO change
			Schedule: "* * * * *", // every 1 hour
			JobTemplate: batchv1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							ServiceAccountName: "gitstar-operator",
							Containers: []corev1.Container{
								{
									Name:  cr.Name + "-gitstarjob",
									Image: "kurisux/gitstar-queryjob:latest",
									Env: []corev1.EnvVar{
										{
											Name:  ENVGitStarName,
											Value: cr.Name,
										},
										{
											Name:  ENVGitStarNameSpace,
											Value: cr.Namespace,
										},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
			ConcurrencyPolicy:          batchv1.ReplaceConcurrent,
			SuccessfulJobsHistoryLimit: &CronJobHistoryLimit,
			FailedJobsHistoryLimit:     &CronJobHistoryLimit,
		},
	}
}

// GenerateCronJobName
func GenerateCronJobName(cr *appv1.GitStar) string {
	return fmt.Sprintf("%s-gitstar", cr.Name)
}
func DeleteCronJob(cr *appv1.GitStar, c client.Client) error {
	return c.Delete(context.TODO(), NewCronJobForCR(cr))
}
