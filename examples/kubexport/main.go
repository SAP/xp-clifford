package main

import (
	"context"
	"log/slog"
	"sort"

	"github.com/SAP/xp-clifford/cli"
	"github.com/SAP/xp-clifford/cli/configparam"
	"github.com/SAP/xp-clifford/cli/export"
	"github.com/SAP/xp-clifford/erratt"
	"github.com/SAP/xp-clifford/mkcontainer"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type collectedResource struct {
	key    string
	object resource.Object
}

func (r *collectedResource) GetGUID() string {
	return string(r.object.GetUID())
}

func (r *collectedResource) GetName() string {
	return r.key
}

type resourceInventory struct {
	ordered []*collectedResource
	index   mkcontainer.TypedContainer[*collectedResource]
}

func newResourceInventory() *resourceInventory {
	return &resourceInventory{
		ordered: make([]*collectedResource, 0),
		index:   mkcontainer.NewTyped[*collectedResource](),
	}
}

func resourceKey(obj resource.Object) string {
	if namespace := obj.GetNamespace(); namespace != "" {
		return obj.GetObjectKind().GroupVersionKind().Kind + "/" + namespace + "/" + obj.GetName()
	}
	return obj.GetObjectKind().GroupVersionKind().Kind + "/" + obj.GetName()
}

func namespaceKey(namespace string) string {
	return "Namespace/" + namespace
}

func (i *resourceInventory) HasKey(key string) bool {
	return len(i.index.GetByName(key)) > 0
}

func (i *resourceInventory) Add(obj resource.Object) bool {
	item := &collectedResource{
		key:    resourceKey(obj),
		object: obj,
	}
	if guid := item.GetGUID(); guid != "" && i.index.GetByGUID(guid) != nil {
		return false
	}
	if i.HasKey(item.GetName()) {
		return false
	}

	i.index.Store(item)
	i.ordered = append(i.ordered, item)
	return true
}

func (i *resourceInventory) Emit(events export.EventHandler) {
	for _, item := range i.ordered {
		events.Resource(item.object)
	}
}

func newClientset() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, erratt.Errorf("cannot load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, erratt.Errorf("cannot create Kubernetes client: %w", err)
	}

	return clientset, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func listNamespaceNames(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, erratt.Errorf("cannot list namespaces for selection: %w", err)
	}

	namespaceNames := make([]string, len(namespaces.Items))
	for i := range namespaces.Items {
		namespaceNames[i] = namespaces.Items[i].GetName()
	}
	sort.Strings(namespaceNames)
	return namespaceNames, nil
}

func resolveNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
	if namespaces := uniqueStrings(namespaceParam.Value()); len(namespaces) > 0 {
		return namespaces, nil
	}

	namespaceNames, err := listNamespaceNames(ctx, clientset)
	if err != nil {
		return nil, err
	}
	if len(namespaceNames) == 0 {
		return nil, erratt.New("cannot select namespaces: no namespaces available")
	}

	namespaceParam.WithPossibleValues(namespaceNames)
	namespaces, err := namespaceParam.ValueOrAsk(ctx)
	if err != nil {
		return nil, erratt.Errorf("cannot determine namespaces: %w", err)
	}
	namespaces = uniqueStrings(namespaces)
	if len(namespaces) == 0 {
		return nil, erratt.New("no namespaces selected for Pod export")
	}

	return namespaces, nil
}

func collectSelectedNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string, inventory *resourceInventory) error {
	if inventory.HasKey(namespaceKey(namespace)) {
		return nil
	}

	namespaceResource, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return erratt.Errorf("cannot get namespace: %w", err).With("namespace", namespace)
	}

	namespaceResource.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Namespace",
	}
	slog.Info("exporting selected namespace", "namespace", namespace)
	inventory.Add(namespaceResource)
	return nil
}

func collectNamespaces(ctx context.Context, clientset *kubernetes.Clientset, inventory *resourceInventory) error {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return erratt.Errorf("cannot list namespaces: %w", err)
	}

	slog.Info("exporting namespaces", "count", len(namespaces.Items))
	for i := range namespaces.Items {
		namespace := namespaces.Items[i].DeepCopy()
		namespace.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		}
		inventory.Add(namespace)
	}

	return nil
}

func collectClusterRoles(ctx context.Context, clientset *kubernetes.Clientset, inventory *resourceInventory) error {
	clusterRoles, err := clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return erratt.Errorf("cannot list cluster roles: %w", err)
	}

	slog.Info("exporting cluster roles", "count", len(clusterRoles.Items))
	for i := range clusterRoles.Items {
		clusterRole := clusterRoles.Items[i].DeepCopy()
		clusterRole.TypeMeta = metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		}
		inventory.Add(clusterRole)
	}

	return nil
}

func collectPods(ctx context.Context, clientset *kubernetes.Clientset, namespace string, inventory *resourceInventory) error {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return erratt.Errorf("cannot list pods: %w", err).With("namespace", namespace)
	}

	slog.Info("exporting pods", "namespace", namespace, "count", len(pods.Items))
	for i := range pods.Items {
		pod := pods.Items[i].DeepCopy()
		pod.TypeMeta = metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		}
		inventory.Add(pod)
	}

	return nil
}

func exportLogic(ctx context.Context, events export.EventHandler) error {
	defer events.Stop()

	kinds, err := export.ResourceKindParam.ValueOrAsk(ctx)
	if err != nil {
		return erratt.Errorf("cannot determine resource kinds: %w", err)
	}
	kinds = uniqueStrings(kinds)
	if len(kinds) == 0 {
		return erratt.New("no resource kinds selected")
	}

	clientset, err := newClientset()
	if err != nil {
		return err
	}

	podNamespaces := []string{}
	for _, kind := range kinds {
		if kind == "Pod" {
			podNamespaces, err = resolveNamespaces(ctx, clientset)
			if err != nil {
				return err
			}
			break
		}
	}

	if len(podNamespaces) > 0 {
		slog.Info("export started", "kind", kinds, "namespaces", podNamespaces)
	} else {
		slog.Info("export started", "kind", kinds)
	}

	inventory := newResourceInventory()

	for _, kind := range kinds {
		switch kind {
		case "Namespace":
			if err := collectNamespaces(ctx, clientset, inventory); err != nil {
				return err
			}
		case "ClusterRole":
			if err := collectClusterRoles(ctx, clientset, inventory); err != nil {
				return err
			}
		case "Pod":
			for _, namespace := range podNamespaces {
				if includeNamespacesParam.Value() {
					if err := collectSelectedNamespace(ctx, clientset, namespace, inventory); err != nil {
						return err
					}
				}
				if err := collectPods(ctx, clientset, namespace, inventory); err != nil {
					return err
				}
			}
		default:
			return erratt.New("unsupported resource kind", "kind", kind)
		}
	}
	inventory.Emit(events)

	return nil
}

var namespaceParam = configparam.StringSlice("namespace", "Namespaces for namespaced resources such as Pod").
	WithShortName("n").
	WithEnvVarName("NAMESPACES")

var includeNamespacesParam = configparam.Bool("include-namespaces", "Also export selected Namespace resources when exporting Pod")

func main() {
	cli.Configuration.ShortName = "kubexport"
	cli.Configuration.ObservedSystem = "Kubernetes"
	export.AddConfigParams(namespaceParam, includeNamespacesParam)
	export.AddResourceKinds("Namespace", "ClusterRole", "Pod")
	export.SetCommand(exportLogic)
	cli.Execute()
}
