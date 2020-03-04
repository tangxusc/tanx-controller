/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	toolv1 "github.com/tangxusc/tanx-controller/api/v1"
)

// NodeJobReconciler reconciles a NodeJob object
type NodeJobReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=tool.tanx.io,resources=nodejobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tool.tanx.io,resources=nodejobs/status,verbs=get;update;patch

func (r *NodeJobReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("nodejob", req.NamespacedName)

	job := &toolv1.NodeJob{}
	err := r.Get(ctx, req.NamespacedName, job)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	job.Spec.PodSpec.RestartPolicy = v1.RestartPolicyOnFailure
	list := &v1.PodList{}
	//创建
	if job.Status.Values == nil || len(job.Status.Values) == 0 {
		list := &v1.NodeList{}
		err = r.List(ctx, list)
		if err != nil {
			return ctrl.Result{}, err
		}
		for _, item := range list.Items {
			err = createPod(ctx, r, job, item)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		job.Status.Finish = false
		err := r.Client.Update(ctx, job)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		//更新状态
		err = r.List(ctx, list, client.InNamespace(req.Namespace), client.MatchingLabels(buildLabels(job)))
		if err != nil {
			return ctrl.Result{}, err
		}
		for _, pod := range list.Items {
			oldstatus := job.Status.Values[pod.Spec.NodeName]
			podPhase := pod.Status.Phase
			if oldstatus != podPhase {
				r.Recorder.Event(job, v1.EventTypeNormal, "ContainerStatusUpdate",
					fmt.Sprintf("Pod[%s].Phase from %v to %v", pod.Name, oldstatus, podPhase))
			}

			job.Status.Values[pod.Spec.NodeName] = podPhase
		}
		allFinish := true
		for _, phase := range job.Status.Values {
			if phase != v1.PodSucceeded {
				allFinish = false
				continue
			}
		}
		job.Status.Finish = allFinish
		if allFinish {
			r.Recorder.Event(job, v1.EventTypeNormal, "JobFinish", fmt.Sprintf("all Pod finished"))
		}
	}
	err = r.Client.Update(ctx, job)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

func createPod(ctx context.Context, r *NodeJobReconciler, job *toolv1.NodeJob, node v1.Node) error {
	name := node.GetName()
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", job.Name, node.Name),
			Namespace: job.Namespace,
			Labels:    buildLabels(job),
		},
		Spec: job.Spec.PodSpec,
	}
	pod.Spec.NodeName = name
	err := controllerutil.SetControllerReference(job, pod, r.Scheme)
	if err != nil {
		return err
	}
	err = r.Create(ctx, pod)
	if err != nil {
		return err
	}
	r.Recorder.Event(job, v1.EventTypeNormal, "ContainerCreate", fmt.Sprintf("create pod[%s] for node[%s]", pod.Name, name))
	if job.Status.Values == nil {
		job.Status.Values = toolv1.NodePodStatus{}
	}
	job.Status.Values[name] = pod.Status.Phase
	return nil
}

func buildLabels(job *toolv1.NodeJob) map[string]string {
	return map[string]string{
		"nodejob-name":   job.Name,
		"nodejobjob-uid": string(job.UID),
	}
}

func (r *NodeJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toolv1.NodeJob{}).
		Complete(r)
}
