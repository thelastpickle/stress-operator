package monitoring

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"text/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GrafanaDashboardKind = "GrafanaDashboard"

type GrafanaTemplateParams struct {
	PrometheusJobName string
	Instance          string
	DashboardName     string
}

func getGrafanaTypes() (schema.GroupVersion, []runtime.Object) {
	gv := schema.GroupVersion{Group: i8ly.SchemeGroupVersion.Group, Version: i8ly.SchemeGroupVersion.Version}
	grafanaTypes := []runtime.Object{&i8ly.GrafanaDashboard{}, &i8ly.GrafanaDashboardList{}, &metav1.ListOptions{}}
	return gv, grafanaTypes
}

func GrafanaDashboardKindExists() (bool, error) {
	return discoveryClient.KindExists(i8ly.SchemeGroupVersion.String(), GrafanaDashboardKind)
}

func GetDashboard(tlpStress *tlp.TLPStress, client client.Client) (*i8ly.GrafanaDashboard, error) {
	dashboard := &i8ly.GrafanaDashboard{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: tlpStress.Namespace, Name: tlpStress.Name}, dashboard)

	return dashboard, err
}

func CreateDashboard(tlpStress *tlp.TLPStress, client client.Client, log logr.Logger) (reconcile.Result, error) {
	dashboard, err := newDashboard(tlpStress.Name, fmt.Sprintf("%s-metrics", tlpStress.Name))
	if err != nil {
		return reconcile.Result{}, err
	}
	dashboard.Name = tlpStress.Name
	dashboard.Namespace = tlpStress.Namespace

	// TODO set controller reference
	err = client.Create(context.TODO(), dashboard)
	if err != nil {
		log.Error(err, "Failed to create dashboard")
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func newDashboard(dashboardName string, prometheusJobName string) (*i8ly.GrafanaDashboard, error) {
	tmpl, err := loadTemplate("stress-dashboard", GrafanaTemplateParams{
		PrometheusJobName: prometheusJobName,
		Instance: dashboardName,
		DashboardName: dashboardName,
	})
	if err != nil {
		return nil, err
	}

	dashboard := i8ly.GrafanaDashboard{}
	err = yaml.Unmarshal(tmpl, &dashboard)

	if err != nil {
		return nil, err
	}

	return &dashboard, nil
}

func loadTemplate(name string, params GrafanaTemplateParams) ([]byte, error) {
	templatePath := os.Getenv("TEMPLATE_PATH")
	path := fmt.Sprintf("%s/%s.yaml", templatePath, name)
	tpl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := template.New("dashboard").Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = parsed.Execute(&buffer, params)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
