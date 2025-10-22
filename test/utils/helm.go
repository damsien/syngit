package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/client-go/rest"
)

func NewDefaultHelmActionConfig(chart Chart) (*action.Configuration, *cli.EnvSettings, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), chart.GetReleaseNamespace(), "secrets", log.Printf)
	return actionConfig, settings, err
}

func NewEnvtestHelmActionConfig(cfg *rest.Config, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	kubeconfig, err := GetKubeconfigFromConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	settings := cli.New()
	tmpKube := filepath.Join(os.TempDir(), "envtest-kubeconfig.yaml")
	if err := os.WriteFile(tmpKube, kubeconfig, 0600); err != nil {
		return nil, settings, err
	}
	settings.KubeConfig = string(kubeconfig)
	settings.RepositoryCache = "/tmp/helmcache"
	settings.RepositoryConfig = "/tmp/helmrepo"

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "secrets", log.Panicf); err != nil {
		return nil, settings, err
	}
	return actionConfig, settings, nil
}

type Chart interface {
	GetChartPath(install *action.Install, settings *cli.EnvSettings) (string, error)
	GetValuesPath() string
	GetChartName() string
	GetChartVersion() string
	GetReleaseName() string
	GetReleaseNamespace() string
}

type BaseChart struct {
	ValuesPath       string
	ChartName        string
	ChartVersion     string
	ReleaseName      string
	ReleaseNamespace string
}

func (chart BaseChart) GetValuesPath() string {
	return chart.ValuesPath
}
func (chart BaseChart) GetChartName() string {
	return chart.ChartName
}
func (chart BaseChart) GetChartVersion() string {
	return chart.ChartVersion
}
func (chart BaseChart) GetReleaseName() string {
	return chart.ReleaseName
}
func (chart BaseChart) GetReleaseNamespace() string {
	return chart.ReleaseNamespace
}

type LocalChart struct {
	BaseChart
	ChartPath string
}

type RemoteChart struct {
	BaseChart
	RepoURL string
}

func (c LocalChart) GetChartPath(install *action.Install, settings *cli.EnvSettings) (string, error) {
	return c.ChartPath + "/" + c.ChartVersion, nil
}

func (c RemoteChart) GetChartPath(install *action.Install, settings *cli.EnvSettings) (string, error) {
	chartPath, err := install.ChartPathOptions.LocateChart(c.ChartName, settings)
	if err != nil {
		return "", err
	}
	return chartPath, nil
}

func InstallChart(chart Chart, actionConfig *action.Configuration, settings *cli.EnvSettings) error {
	install := action.NewInstall(actionConfig)
	install.ReleaseName = chart.GetReleaseName()
	install.Namespace = chart.GetReleaseNamespace()
	install.CreateNamespace = true
	install.Timeout = time.Minute * 10
	install.Wait = true
	install.WaitForJobs = true

	chartPath, err := chart.GetChartPath(install, settings)
	if err != nil {
		return err
	}

	valsOpt := &values.Options{
		ValueFiles: []string{chart.GetValuesPath()},
	}
	vals, err := valsOpt.MergeValues(getter.All(settings))
	if err != nil {
		return err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	// if chartRequested.Metadata.Dependencies != nil {
	// 	if err := action.CheckDependencies(chartRequested, chartRequested.Metadata.Dependencies); err != nil {
	// 		// maybe handle dependency update via downloader.Manager
	// 		dm := &downloader.Manager{
	// 			ChartPath:        chartPath,
	// 			Keyring:          install.ChartPathOptions.Keyring,
	// 			SkipUpdate:       false,
	// 			Getters:          getter.All(settings),
	// 			RepositoryConfig: settings.RepositoryConfig,
	// 			RepositoryCache:  settings.RepositoryCache,
	// 			RegistryClient:   install.GetRegistryClient(),
	// 		}
	// 		if err := dm.Update(); err != nil {
	// 			return err
	// 		}
	// 		// reload the chart
	// 		chartRequested, err = loader.Load(chartPath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	_, err = install.Run(chartRequested, vals)
	if err != nil {
		return err
	}

	return nil
}

func UninstallChart(chart Chart, actionConfig *action.Configuration, settings *cli.EnvSettings) error {
	uninstall := action.NewUninstall(actionConfig)
	uninstall.KeepHistory = false

	_, err := uninstall.Run(chart.GetReleaseName())
	if err != nil {
		return err
	}

	return nil
}

// GetLatestChartVersion returns the latest version from the charts directory
func GetLatestChartVersion(chartsDir string) (string, error) {
	// Read all directories in charts/
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return "", err
	}

	// Filter directories and get their names
	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] < versions[j]
	})

	// Return the latest version or empty if no versions found
	if len(versions) == 0 {
		return "", fmt.Errorf("no chart versions found in %s", chartsDir)
	}

	return versions[len(versions)-1], nil
}
