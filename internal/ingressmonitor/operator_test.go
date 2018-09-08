package ingressmonitor

import (
	"errors"
	"testing"
	"time"

	"github.com/jelmersnoeck/ingress-monitor/apis/ingressmonitor/v1alpha1"
	"github.com/jelmersnoeck/ingress-monitor/internal/provider"
	"github.com/jelmersnoeck/ingress-monitor/internal/provider/fake"
	"github.com/jelmersnoeck/ingress-monitor/pkg/client/generated/clientset/versioned"
	imfake "github.com/jelmersnoeck/ingress-monitor/pkg/client/generated/clientset/versioned/fake"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func namespaceKey(t *testing.T, obj interface{}) string {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		t.Fatalf("Could not get NamespaceKey for object %#v", obj)
	}

	return key
}

func TestOperator_HandleIngressMonitor(t *testing.T) {
	t.Run("creating", func(t *testing.T) {
		t.Run("without registered provider", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "test",
						},
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			expErr := errors.New("Error fetching provider 'test': the specified provider can't be found")
			if err := op.handleIngressMonitor(namespaceKey(t, crd)); err.Error() != expErr.Error() {
				t.Errorf("Expected '%s' error, got '%s'", expErr, err)
			}
		})

		t.Run("with error creating monitor", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			err := errors.New("my-provider-error")
			prov := new(fake.SimpleProvider)
			prov.CreateFunc = func(v1alpha1.MonitorTemplateSpec) (string, error) {
				return "", err
			}

			fact.Register("simple", fake.FactoryFunc(prov))

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "simple",
						},
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			if handleErr := op.handleIngressMonitor(namespaceKey(t, crd)); err != handleErr {
				t.Errorf("Expected error '%s', got %s", err, handleErr)
			}

			if prov.CreateCount != 1 {
				t.Errorf("Expected Create to be called once, got %d", prov.CreateCount)
			}
		})

		t.Run("without errors", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			prov := new(fake.SimpleProvider)
			prov.CreateFunc = func(v1alpha1.MonitorTemplateSpec) (string, error) {
				return "1234", nil
			}

			fact.Register("simple", fake.FactoryFunc(prov))

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "simple",
						},
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			if err := op.handleIngressMonitor(namespaceKey(t, crd)); err != nil {
				t.Errorf("Expected no error, got %s", err)
			}

			if prov.CreateCount != 1 {
				t.Errorf("Expected Create to be called once, got %d", prov.CreateCount)
			}

			crd, err := crdClient.Ingressmonitor().IngressMonitors(crd.Namespace).Get(crd.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Expected no error fetching the CRD, got %s", err)
			}

			if crd.Status.ID != "1234" {
				t.Errorf("Expected status to be updated")
			}
		})

		t.Run("with ID already set", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			prov := new(fake.SimpleProvider)
			fact.Register("simple", fake.FactoryFunc(prov))

			prov.UpdateFunc = func(status string, _ v1alpha1.MonitorTemplateSpec) (string, error) {
				if status != "1234" {
					t.Errorf("Expected status to be `1234`, got `%s`", status)
				}
				return status, nil
			}

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "simple",
						},
					},
				},
				Status: v1alpha1.IngressMonitorStatus{
					ID: "1234",
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			if err := op.handleIngressMonitor(namespaceKey(t, crd)); err != nil {
				t.Errorf("Expected no error, got %s", err)
			}

			if prov.CreateCount != 0 {
				t.Errorf("Did not expect an object to be created, got a create call")
			}

			if prov.UpdateCount != 1 {
				t.Errorf("Expected provider to do one update, got %d", prov.UpdateCount)
			}
		})
	})

	t.Run("updating", func(t *testing.T) {
		t.Run("without registered provider", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			prov := new(fake.SimpleProvider)
			fact.Register("simple", fake.FactoryFunc(prov))

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "test",
						},
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			expErr := errors.New("Error fetching provider 'test': the specified provider can't be found")
			if err := op.handleIngressMonitor(namespaceKey(t, crd)); err.Error() != expErr.Error() {
				t.Errorf("Expected error '%s', got '%s'", expErr, err)
			}

			if prov.UpdateCount != 0 {
				t.Errorf("Expected no updates, got %d", prov.UpdateCount)
			}

			if prov.CreateCount != 0 {
				t.Errorf("Expected no updates, got %d", prov.UpdateCount)
			}
		})

		t.Run("with error updating monitor", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			err := errors.New("my-provider-error")
			prov := new(fake.SimpleProvider)
			prov.UpdateFunc = func(status string, _ v1alpha1.MonitorTemplateSpec) (string, error) {
				if status != "12345" {
					t.Errorf("Expected status to be `12345`, got `%s`", status)
				}
				return status, err
			}

			fact.Register("simple", fake.FactoryFunc(prov))

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "simple",
						},
					},
				},
				Status: v1alpha1.IngressMonitorStatus{
					ID: "12345",
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			if handleErr := op.handleIngressMonitor(namespaceKey(t, crd)); err != handleErr {
				t.Errorf("Expected error '%s', got '%s'", err, handleErr)
			}

			if prov.UpdateCount != 1 {
				t.Errorf("Expected Update to be called once, got %d", prov.UpdateCount)
			}
		})

		t.Run("without errors", func(t *testing.T) {
			fact := provider.NewFactory(nil)

			prov := new(fake.SimpleProvider)
			prov.UpdateFunc = func(status string, _ v1alpha1.MonitorTemplateSpec) (string, error) {
				if status != "12345" {
					t.Errorf("Expected status to be `12345`, got `%s`", status)
				}
				return status, nil
			}

			fact.Register("simple", fake.FactoryFunc(prov))

			crd := &v1alpha1.IngressMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-im",
					Namespace: "testing",
				},
				Spec: v1alpha1.IngressMonitorSpec{
					Provider: v1alpha1.NamespacedProvider{
						Namespace: "testing",
						ProviderSpec: v1alpha1.ProviderSpec{
							Type: "simple",
						},
					},
				},
				Status: v1alpha1.IngressMonitorStatus{
					ID: "12345",
				},
			}

			crdClient := imfake.NewSimpleClientset(crd)
			op, _ := NewOperator(nil, crdClient, "", time.Minute, fact)

			if err := op.handleIngressMonitor(namespaceKey(t, crd)); err != nil {
				t.Errorf("Expected no error, got %s", err)
			}

			if prov.UpdateCount != 1 {
				t.Errorf("Expected Update to be called once, got %d", prov.UpdateCount)
			}
		})
	})
}

func TestOperator_OnDelete_IngressMonitor(t *testing.T) {
	t.Run("without registered provider", func(t *testing.T) {
		fact := provider.NewFactory(nil)
		op, _ := NewOperator(nil, nil, "", time.Minute, fact)

		prov := new(fake.SimpleProvider)
		fact.Register("simple", fake.FactoryFunc(prov))

		crd := &v1alpha1.IngressMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-im",
				Namespace: "testing",
			},
			Spec: v1alpha1.IngressMonitorSpec{
				Provider: v1alpha1.NamespacedProvider{
					Namespace: "testing",
					ProviderSpec: v1alpha1.ProviderSpec{
						Type: "test",
					},
				},
			},
		}

		op.OnDelete(crd)

		if prov.DeleteCount != 0 {
			t.Errorf("Expected no deletes, got %d", prov.DeleteCount)
		}
	})

	t.Run("with error deleting monitor", func(t *testing.T) {
		fact := provider.NewFactory(nil)

		err := errors.New("my-provider-error")
		prov := new(fake.SimpleProvider)
		prov.DeleteFunc = func(status string) error {
			if status != "12345" {
				t.Errorf("Expected status to be `12345`, got `%s`", status)
			}
			return err
		}

		fact.Register("simple", fake.FactoryFunc(prov))
		op, _ := NewOperator(nil, nil, "", time.Minute, fact)

		crd := &v1alpha1.IngressMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-im",
				Namespace: "testing",
			},
			Spec: v1alpha1.IngressMonitorSpec{
				Provider: v1alpha1.NamespacedProvider{
					Namespace: "testing",
					ProviderSpec: v1alpha1.ProviderSpec{
						Type: "simple",
					},
				},
			},
			Status: v1alpha1.IngressMonitorStatus{
				ID: "12345",
			},
		}

		op.OnDelete(crd)

		if prov.DeleteCount != 1 {
			t.Errorf("Expected delete to be called once, got %d", prov.DeleteCount)
		}
	})

	t.Run("without errors", func(t *testing.T) {
		fact := provider.NewFactory(nil)

		prov := new(fake.SimpleProvider)
		prov.DeleteFunc = func(status string) error {
			if status != "12345" {
				t.Errorf("Expected status to be `12345`, got `%s`", status)
			}
			return nil
		}

		fact.Register("simple", fake.FactoryFunc(prov))

		crd := &v1alpha1.IngressMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-im",
				Namespace: "testing",
			},
			Spec: v1alpha1.IngressMonitorSpec{
				Provider: v1alpha1.NamespacedProvider{
					Namespace: "testing",
					ProviderSpec: v1alpha1.ProviderSpec{
						Type: "simple",
					},
				},
			},
			Status: v1alpha1.IngressMonitorStatus{
				ID: "12345",
			},
		}
		op, _ := NewOperator(nil, nil, "", time.Minute, fact)

		op.OnDelete(crd)

		if prov.DeleteCount != 1 {
			t.Errorf("Expected Update to be called once, got %d", prov.DeleteCount)
		}
	})
}

func Test_OnAdd_Monitor(t *testing.T) {
	t.Run("without matching ingresses", func(t *testing.T) {
		ing := &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "testing",
				Labels: map[string]string{
					"service": "no-match",
				},
			},
		}

		k8sClient := k8sfake.NewSimpleClientset(ing)
		crdClient := imfake.NewSimpleClientset()

		fact := provider.NewFactory(nil)
		op, _ := NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, fact)

		mon := &v1alpha1.Monitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-monitor",
				Namespace: "testing",
			},
			Spec: v1alpha1.MonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"team": "gophers",
					},
				},
			},
		}

		op.OnAdd(mon)

		imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).
			List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Could not get IngressMonitor List: %s", err)
		}

		if len(imList.Items) != 0 {
			t.Errorf("Expected no IngressMonitors to be registered, got %d", len(imList.Items))
		}
	})

	t.Run("with matching ingresses", func(t *testing.T) {
		ing := &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "testing",
				Labels: map[string]string{
					"team": "gophers",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{Host: "test-host.sphc.io"},
					{Host: "test-app.sphc.io"},
				},
			},
		}

		k8sClient := k8sfake.NewSimpleClientset(ing)

		t.Run("without provider configured", func(t *testing.T) {
			crdClient := imfake.NewSimpleClientset()
			op, _ := NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, provider.NewFactory(nil))

			mon := &v1alpha1.Monitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-monitor",
					Namespace: "testing",
				},
				Spec: v1alpha1.MonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"team": "gophers",
						},
					},
					Provider: v1.LocalObjectReference{
						Name: "test-provider",
					},
				},
			}

			op.OnAdd(mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).
				List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Could not get IngressMonitor List: %s", err)
			}

			if len(imList.Items) != 0 {
				t.Errorf("Expected 0 IngressMonitor to be registered, got %d", len(imList.Items))
			}
		})

		t.Run("with everything configured", func(t *testing.T) {
			prov := &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "testing",
				},
			}

			tmpl := &v1alpha1.MonitorTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "testing",
				},
				Spec: v1alpha1.MonitorTemplateSpec{
					Type: "HTTP",
					HTTP: &v1alpha1.HTTPTemplate{
						Endpoint: ptrString("/_healthz"),
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(prov, tmpl)
			op, _ := NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, provider.NewFactory(nil))

			mon := &v1alpha1.Monitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-monitor",
					Namespace: "testing",
				},
				Spec: v1alpha1.MonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"team": "gophers",
						},
					},
					Provider: v1.LocalObjectReference{
						Name: "test-provider",
					},
					Template: v1.LocalObjectReference{
						Name: "test-template",
					},
				},
			}

			op.OnAdd(mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).
				List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Could not get IngressMonitor List: %s", err)
			}

			if len(imList.Items) != 2 {
				t.Errorf("Expected 2 IngressMonitor to be registered, got %d", len(imList.Items))
			}
		})

		t.Run("with templating set up", func(t *testing.T) {
			prov := &v1alpha1.Provider{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-provider",
					Namespace: "testing",
				},
			}

			tmpl := &v1alpha1.MonitorTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "testing",
				},
				Spec: v1alpha1.MonitorTemplateSpec{
					Name: "some-test-{{.IngressName}}-{{.IngressNamespace}}",
					Type: "HTTP",
					HTTP: &v1alpha1.HTTPTemplate{
						Endpoint: ptrString("/_healthz"),
					},
				},
			}

			crdClient := imfake.NewSimpleClientset(prov, tmpl)
			op, _ := NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, provider.NewFactory(nil))

			mon := &v1alpha1.Monitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-monitor",
					Namespace: "testing",
				},
				Spec: v1alpha1.MonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"team": "gophers",
						},
					},
					Provider: v1.LocalObjectReference{
						Name: "test-provider",
					},
					Template: v1.LocalObjectReference{
						Name: "test-template",
					},
				},
			}

			op.OnAdd(mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).
				List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Could not get IngressMonitor List: %s", err)
			}

			if len(imList.Items) != 2 {
				t.Errorf("Expected 2 IngressMonitor to be registered, got %d", len(imList.Items))
			}

			// check if the templated name is parsed
			expectedName := "some-test-test-ingress-testing"
			if name := imList.Items[0].Spec.Template.Name; name != expectedName {
				t.Errorf("Expected name to be `%s`, got `%s", expectedName, name)
			}
		})
	})

	t.Run("it should set up values correctly", func(t *testing.T) {
		ing := &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress",
				Namespace: "testing",
				Labels: map[string]string{
					"team": "gophers",
				},
			},
			Spec: v1beta1.IngressSpec{
				TLS: []v1beta1.IngressTLS{
					{
						Hosts: []string{
							"test-host.sphc.io",
						},
					},
				},
				Rules: []v1beta1.IngressRule{
					{Host: "test-host.sphc.io"},
				},
			},
		}

		k8sClient := k8sfake.NewSimpleClientset(ing)

		prov := &v1alpha1.Provider{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-provider",
				Namespace: "testing",
			},
		}

		tmpl := &v1alpha1.MonitorTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-template",
				Namespace: "testing",
			},
			Spec: v1alpha1.MonitorTemplateSpec{
				Type: "HTTP",
				HTTP: &v1alpha1.HTTPTemplate{
					Endpoint: ptrString("/_healthz"),
				},
			},
		}

		crdClient := imfake.NewSimpleClientset(prov, tmpl)
		op, _ := NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, provider.NewFactory(nil))

		mon := &v1alpha1.Monitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-monitor",
				Namespace: "testing",
			},
			Spec: v1alpha1.MonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"team": "gophers",
					},
				},
				Provider: v1.LocalObjectReference{
					Name: "test-provider",
				},
				Template: v1.LocalObjectReference{
					Name: "test-template",
				},
			},
		}

		op.OnAdd(mon)

		imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).
			List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Could not get IngressMonitor List: %s", err)
		}

		if len(imList.Items) != 1 {
			t.Errorf("Expected 1 IngressMonitor, got %d", len(imList.Items))
		}

		im := imList.Items[0]
		expURL := "https://test-host.sphc.io/_healthz"
		if url := im.Spec.Template.HTTP.URL; url != expURL {
			t.Errorf("Expected URL to be `%s`, got `%s`", expURL, url)
		}
	})
}

func Test_OnUpdate_Monitor(t *testing.T) {
	ing1 := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "go-ingress",
			Namespace: "testing",
			Labels: map[string]string{
				"team":  "gophers",
				"squad": "operations",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{Host: "api.example.com"},
			},
		},
	}
	ing2 := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-ingress",
			Namespace: "testing",
			Labels: map[string]string{
				"team": "reacters",
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{Host: "api.foo.com"},
			},
		},
	}

	prov := &v1alpha1.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-provider",
			Namespace: "testing",
		},
	}
	tpl := &v1alpha1.MonitorTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-template",
			Namespace: "testing",
		},
		Spec: v1alpha1.MonitorTemplateSpec{
			Type: "HTTP",
			HTTP: &v1alpha1.HTTPTemplate{},
		},
	}

	mon := &v1alpha1.Monitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-monitor",
			Namespace: "testing",
		},
		Spec: v1alpha1.MonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"team": "gophers",
				},
			},
			Provider: v1.LocalObjectReference{
				Name: "test-provider",
			},
			Template: v1.LocalObjectReference{
				Name: "test-template",
			},
		},
	}

	var crdClient versioned.Interface
	var k8sClient kubernetes.Interface
	var op *Operator

	setup := func() {
		crdClient = imfake.NewSimpleClientset(prov, tpl)
		mon, _ = crdClient.Ingressmonitor().Monitors(mon.Namespace).Create(mon)
		k8sClient = k8sfake.NewSimpleClientset()
		ing1, _ = k8sClient.Extensions().Ingresses(ing1.Namespace).Create(ing1)
		ing2, _ = k8sClient.Extensions().Ingresses(ing2.Namespace).Create(ing2)

		op, _ = NewOperator(k8sClient, crdClient, v1.NamespaceAll, time.Minute, provider.NewFactory(nil))
		// we won't start the operator so the informers aren't automatically
		// trigerred. Make sure the monitor is added correctly.
		op.OnAdd(mon)
	}

	t.Run("without CRD changes", func(t *testing.T) {
		t.Run("without Ingress changes", func(t *testing.T) {
			setup()
			op.OnUpdate(mon, mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 1 {
				t.Errorf("Expected 1 IngressMonitor, got %d", len(imList.Items))
			}

			if !metav1.IsControlledBy(&imList.Items[0], ing1) {
				t.Errorf("Expected IngressMonitor to be owned by the correct Ingress")
			}
		})

		t.Run("with Ingress changes", func(t *testing.T) {
			setup()

			// the user has changed the Ingress and removed the labels in their
			// manifest
			ing := ing1.DeepCopy()
			ing.Labels = map[string]string{}
			k8sClient.Extensions().Ingresses(ing.Namespace).Update(ing)

			op.OnUpdate(mon, mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 0 {
				t.Errorf("Expected 0 IngressMonitor, got %d", len(imList.Items))
			}
		})

		t.Run("with Ingress additions", func(t *testing.T) {
			setup()

			// the user has changed the Ingress and removed the labels in their
			// manifest
			ing := ing2.DeepCopy()
			ing.Labels["team"] = "gophers"
			k8sClient.Extensions().Ingresses(ing.Namespace).Update(ing)

			op.OnUpdate(mon, mon)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 2 {
				t.Errorf("Expected 2 IngressMonitors, got %d", len(imList.Items))
			}
		})
	})

	t.Run("with CRD changes", func(t *testing.T) {
		t.Run("to change Ingresses", func(t *testing.T) {
			setup()

			new := mon.DeepCopy()
			new.Spec.Selector.MatchLabels["team"] = "reacters"
			op.OnUpdate(mon, new)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 1 {
				t.Errorf("Expected 1 IngressMonitor, got %d", len(imList.Items))
			}

			if !metav1.IsControlledBy(&imList.Items[0], ing2) {
				t.Errorf("Expected IngressMonitor to be owned by the correct Ingress")
			}
		})

		t.Run("with the same Ingress", func(t *testing.T) {
			setup()

			new := mon.DeepCopy()
			// add a new label which makes selection more specific
			new.Spec.Selector.MatchLabels["squad"] = "operations"
			op.OnUpdate(mon, new)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 1 {
				t.Errorf("Expected 1 IngressMonitor, got %d", len(imList.Items))
			}

			if !metav1.IsControlledBy(&imList.Items[0], ing1) {
				t.Errorf("Expected IngressMonitor to be owned by the correct Ingress")
			}
		})

		t.Run("without matching Ingress", func(t *testing.T) {
			setup()

			new := mon.DeepCopy()
			// no ingress is set up with this label
			new.Spec.Selector.MatchLabels["non"] = "existing"
			op.OnUpdate(mon, new)

			imList, err := crdClient.Ingressmonitor().IngressMonitors(mon.Namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Expected no error listing IngressMonitors, got %s", err)
			}

			if len(imList.Items) != 0 {
				t.Errorf("Expected 0 IngressMonitors, got %d", len(imList.Items))
			}
		})
	})
}

func ptrString(s string) *string {
	return &s
}
