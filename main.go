package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	options "github.com/mreiferson/go-options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

var opts = NewOptions()

// WaitPodsToBeReady waits until all pods be ready before return.
func WaitPodsToBeReady(c *kubernetes.Clientset) {
	stopCh := make(chan struct{})
	wait.Until(func() {
		// this gives us a few quick retries before a long pause and then a few more quick retries
		err := wait.ExponentialBackoff(retry.DefaultRetry, func() (bool, error) {

			pods, err := c.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: ""})
			if err != nil {
				LogError("Error getting Pods", err)
				return false, nil
			}
			fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

			ready := true
			for _, p := range pods.Items {
				if p.Status.Phase != "Running" {
					ready = false
				}
			}

			if ready {
				fmt.Println("All pods running")
				return true, nil
			}

			return false, err
		})
		if err == nil {
			close(stopCh)
		}
	}, 1*time.Minute, stopCh)
}

func SetupCoreAddons() {
	chartsPath := fmt.Sprintf("%s/charts", opts.AssetsPath)

	charts, err := ioutil.ReadDir(chartsPath)
	if err != nil {
		LogError("cannot list charts", err)
	}

	for _, c := range charts {
		fmt.Printf("installing chart %s\n", c.Name())

		cPath := fmt.Sprintf("%s/%s/", chartsPath, c.Name())
		cmd := exec.Command("helm", "install", "--host=127.0.0.1:44134", "--name", c.Name(), cPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := fmt.Sprintf("failed to install %s", c)
			LogError(msg, err)
		} else {
			fmt.Printf(string(out))
		}
	}

	// install tiller
	cmd := exec.Command("helm", "init", "--node-selectors node-role.kubernetes.io/master=''",
		"--override", "'spec.template.spec.tolerations[0].key'='node-role.kubernetes.io/master'",
		"--override", "'spec.template.spec.tolerations[0].operator'='Exists'",
		"--override", "'spec.template.spec.tolerations[0].effect'='NoSchedule'")
	out, err := cmd.CombinedOutput()
	if err != nil {
		LogError("failed to install tiller", err)
	} else {
		fmt.Printf(string(out))
	}
}

func BootstrapCluster() {
	tlsPath := fmt.Sprintf("%s/tls", opts.AssetsPath)
	authPath := fmt.Sprintf("%s/auth", opts.AssetsPath)
	controlplaneManifestsPath := fmt.Sprintf("%s/bootstrap-manifests", opts.AssetsPath)
	bootstrapSecretsPath := fmt.Sprintf("%s/bootstrap-secrets", opts.KubernetesPath)
	manifestsPath := fmt.Sprintf("%s/manifests", opts.KubernetesPath)

	// copying certificates
	err := CopyDir(tlsPath, bootstrapSecretsPath)
	if err != nil {
		LogError("something went wrong while copying certificates", err)
	}

	// copyting auth file
	err = CopyDir(authPath, bootstrapSecretsPath)
	if err != nil {
		LogError("something went wrong while copying kubeconfig", err)
	}

	// copying kube-apiserver static manifest
	err = CopyDir(controlplaneManifestsPath, manifestsPath)
	if err != nil {
		LogError("something went wrong while copying control plane manifest", err)
	}
}

func main() {
	flagSet := flag.NewFlagSet("k8s-cluster-manager", flag.ExitOnError)
	config := flagSet.String("config", "", "path to config file")
	showVersion := flagSet.Bool("version", false, "print version string")

	flagSet.String("assets", "/app/assets", "assets path")
	flagSet.String("kubernetes", "/app/kubernetes", "host path")
	flagSet.String("kubeconfig", "", "host path")

	flagSet.Parse(os.Args[1:])

	if *showVersion {
		fmt.Printf("k8s-cluster-manager v%s (rev:%s built with %s)\n", VERSION, GitCommit, runtime.Version())
		return
	}

	cfg := make(EnvOptions)
	if *config != "" {
		_, err := toml.DecodeFile(*config, &cfg)
		if err != nil {
			log.Fatalf("ERROR: failed to load config file %s - %s", *config, err)
		}
	}
	cfg.LoadEnvForStruct(opts)
	options.Resolve(opts, flagSet, cfg)
	err := opts.Validate()
	if err != nil {
		LogError("error validating configuration", err)
		os.Exit(1)
	}

	// use the current context in kubeconfig
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", opts.KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	BootstrapCluster()

	WaitPodsToBeReady(clientset)

	SetupCoreAddons()

}
