package tenant

import (
	"context"

	tenantsv1alpha1 "github.com/example-inc/tenant-operator/pkg/apis/tenants/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_tenant")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Tenant Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTenant{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tenant-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Tenant
	err = c.Watch(&source.Kind{Type: &tenantsv1alpha1.Tenant{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Tenant
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tenantsv1alpha1.Tenant{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTenant{}

// ReconcileTenant reconciles a Tenant object
type ReconcileTenant struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Tenant object and makes changes based on the state read
// and what is in the Tenant.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTenant) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Tenant")

	// Fetch the Tenant instance
	instance := &tenantsv1alpha1.Tenant{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Namespace object
        // TODO: Handle multiple namespaces inside a tenant
	namespace := newNamespaceForCR(instance)

	// Set Tenant instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, namespace, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

        var ns_name = instance.Spec.Namespace
	// Check if this Namespace already exists
	found := &corev1.Namespace{}
        // TODO: need to check if namespace exists instead of CR
	err = r.client.Get(context.TODO(), request.NamespacedName, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", ns_name)
		err = r.client.Create(context.TODO(), namespace)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Namespace created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Namespace already exists - don't requeue
	reqLogger.Info("Skip reconcile: Namespace already exists",  found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newNamespaceForCR(cr *tenantsv1alpha1.Tenant) *corev1.Namespace {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Namespace ,
			Labels:    labels,
		},
	}
}
