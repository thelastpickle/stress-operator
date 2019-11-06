package monitoring

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	tlp "github.com/jsanda/tlp-stress-operator/pkg/apis/thelastpickle/v1alpha1"
	"github.com/jsanda/tlp-stress-operator/pkg/k8s"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"text/template"
	"time"
)

const (
	GrafanaKind          = "Grafana"
	GrafanaDashboardKind = "GrafanaDashboard"
	GrafanaName          = "tlpstress-grafana"
	DataSourceName       = PrometheusName
)

type GrafanaTemplateParams struct {
	PrometheusJobName string
	Instance          string
	DashboardName     string
	DataSource        string
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
	dashboard.ObjectMeta.Labels = map[string]string{
		"app": "tlpstress",
	}

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
		DataSource: DataSourceName,
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

func GrafanaKindExists() (bool, error) {
	return discoveryClient.KindExists(i8ly.SchemeGroupVersion.String(), GrafanaKind)
}

func GetGrafana(namespace string, client client.Client) (*i8ly.Grafana, error) {
	instance := &i8ly.Grafana{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: GrafanaName}, instance)

	return instance, err
}

func CreateGrafana(namespace string, client client.Client, log logr.Logger) (reconcile.Result, error) {
	instance := newGrafana(namespace)
	log.Info("Creating Grafana", "Grafana.Namespace", instance.Namespace, "Grafana.Name",
		instance.Name)
	if err := k8s.CreateResource(client, instance); err != nil {
		log.Error(err, "Failed to create Grafana", "Grafana.Namespace", instance.Namespace,
			"Grafana.Name", instance.Name)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
}

func newGrafana(namespace string) *i8ly.Grafana {
	selector := &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key: "app",
				Operator: metav1.LabelSelectorOpIn,
				Values: []string{"tlpstress"},
			},
		},
	}

	return &i8ly.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      GrafanaName,
		},
		Spec: i8ly.GrafanaSpec{
			Service: i8ly.GrafanaService{
				Labels: map[string]string{
					"app": "grafana",
				},
			},
			Config: i8ly.GrafanaConfig{
				Log: i8ly.GrafanaConfigLog{
					Mode: "console",
					Level: "debug",
				},
				Security: i8ly.GrafanaConfigSecurity{
					AdminUser: "root",
					AdminPassword: "grafana",
				},
				Auth: i8ly.GrafanaConfigAuth{
					DisableLoginForm: false,
					DisableSignoutMenu: false,
				},
				AuthBasic: i8ly.GrafanaConfigAuthBasic{
					Enabled: true,
				},
				AuthAnonymous: i8ly.GrafanaConfigAuthAnonymous{
					Enabled: true,
				},
			},
			DashboardLabelSelector: []*metav1.LabelSelector{selector},
		},
	}
}

func GetDataSource(namespace string, client client.Client) (*i8ly.GrafanaDataSource, error) {
	ds := &i8ly.GrafanaDataSource{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: DataSourceName}, ds)

	return ds, err
}

func CreateDataSource(namespace string, client client.Client, log logr.Logger) (reconcile.Result, error) {
	log.Info("Creating Prometheus data source", "GrafanaDataSource.Namespace", namespace,
		"GrafanaDataSource.Name", PrometheusName)

	ds := &i8ly.GrafanaDataSource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name: PrometheusName,
		},
		Spec: i8ly.GrafanaDataSourceSpec{
			Name: "middleware.yaml",
			Datasources: []i8ly.GrafanaDataSourceFields{
				{
					Name: PrometheusName,
					Type: "prometheus",
					Access: "proxy",
					Url: "http://tlpstress-prometheus:9090",
					IsDefault: true,
					Version: 1,
					JsonData: i8ly.GrafanaDataSourceJsonData{
						TlsSkipVerify: true,
						TimeInterval: "5s",
					},
				},
			},
		},
	}
	if err := k8s.CreateResource(client, ds); err != nil {
		log.Error(err, "Failed to create Prometheus data source", "GrafanaDataSource.Namespace", namespace,
			"GrafanaDataSource.Name", PrometheusName)
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
}
