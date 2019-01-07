package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	options "github.com/mreiferson/go-options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var opts = NewOptions()

type Status struct {
	IsAPIRunning        bool
	IsBootstrapFinished bool
}

// WaitPodsToBeReady waits until all pods be ready before return.
// TODO: implement exponentional backoffs and eventually fails
func WaitPodsToBeReady(c *kubernetes.Clientset) {
	for {
		pods, err := c.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: ""})
		if err != nil {
			panic(err.Error())
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
			return
		}

		time.Sleep(10 * time.Second)
	}
}

func SetupCoreAddons() {
	chartsPath := fmt.Sprintf("%s/charts", opts.AssetsPath)

	// cmd := exec.Command("helm", "template", chartsPath, fmt.Sprintf("--output %s/_generated/", chartsPath))

	charts, err := ioutil.ReadDir(chartsPath)
	if err != nil {
		LogError("cannot list charts", err)
	}

	for _, c := range charts {
		fmt.Printf("installing chart %s\n", c.Name())

		cPath := fmt.Sprintf("%s/%s/", chartsPath, c.Name())
		cmd := exec.Command("helm", "install", cPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			LogError("cmd.Run() failed with", err)
		} else {
			fmt.Printf(string(out))
		}
	}

	// decode := scheme.Codecs.UniversalDeserializer().Decode
	// obj, _, err := decode(out, nil, nil)

	// if err != nil {
	// 	log.Fatal(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
	// }

	// fmt.Printf("%v", obj)

	// now use switch over the type of the object
	// and match each type-case
	// switch o := obj.(type) {
	// case *v1.Pod:
	// 	// o is a pod
	// case *v1beta1.Role:
	// 	// o is the actual role Object with all fields etc
	// case *v1beta1.RoleBinding:
	// case *v1beta1.ClusterRole:
	// case *v1beta1.ClusterRoleBinding:
	// case *v1.ServiceAccount:
	// default:
	// 	//o is unknown for us
	// }

	// fmt.Printf(string(out))
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
	flagSet.String("failover-ips", "10.4.0.1,10.4.0.2", "failover-ips of API server")

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

	// connection to k8s
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}
