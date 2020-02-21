package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/task"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	"os"
	"os/exec"
	"time"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/rbac"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy torpedo to a k8s cluster",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if chaosMesh {
			out, err := exec.Command("kubectl", "apply", "-f", "manifests/").Output()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(out))
			t := func() (interface{}, bool, error) {
				if _, err = exec.Command("kubectl", "get", "crd", "podchaos.pingcap.com").Output(); err != nil {
					return nil, true, err
				}
				return nil, false, nil
			}
			if _, err = task.DoRetryWithTimeout(t, 1*time.Minute, 5*time.Second); err != nil {
				fmt.Printf("failed to create CRD for chaosmesh Cause: %v", err)
				os.Exit(1)
			}

			out, err = exec.Command("kubectl", "create", "namespace", "chaos-testing").Output()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(out))

			out, err = exec.Command("helm", "install", "chaos-mesh", "chaos-mesh", "--namespace=chaos-testing").Output()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(out))
			dep := &appsv1.Deployment{}
			dep.Name = "chaos-controller-manager"
			dep.Namespace = "chaos-testing"
			if err = apps.Instance().ValidateDeployment(dep, 3*time.Minute, 10*time.Second); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err = apps.Instance().ValidateDaemonSet("chaos-daemon", "chaos-testing", 3*time.Minute); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			out, err = exec.Command("kubectl", "apply", "-f", "../drivers/scheduler/k8s/specs/chaos/").Output()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		createTorpedo()
	},
}

var (
	verbose                    bool
	scaleFactor                int
	scheduler                  string
	loglevel                   string
	chaosLevel                 int
	minRunTime                 int
	upgradeEndpointUrl         string
	upgradeEndpointVersion     string
	provisioner                string
	storageDriver              string
	configmap                  string
	failFast                   bool
	skipTests                  string
	focusTests                 string
	torpedoImg                 string
	timeout                    time.Duration
	appDestroyTimeout          time.Duration
	driverStartTimeout         time.Duration
	storagenodeRecoveryTimeout time.Duration
	azureTenantID              string
	azureClientID              string
	azureClientSecret          string
	testSuite                  string
	detach                     bool
	torpedoSSHKey              string
	torpedoSSHUser             string
	torpedoSSHPassword         string
	customAppConfig            string
	appList                    string
	nodeDriver                 string
	chaosMesh                  bool
)

func createTorpedo() {
	serviceAccount := corev1.ServiceAccount{}
	serviceAccount.Name = "torpedo-account"

	clusterRole := rbacv1.ClusterRole{}
	clusterRole.Name = "torpedo-role"
	clusterRole.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
		{
			NonResourceURLs: []string{"*"},
			Verbs:           []string{"*"},
		},
	}

	clusterRoleBinding := rbacv1.ClusterRoleBinding{}
	clusterRoleBinding.Name = "torpedo-role-binding"
	clusterRoleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "torpedo-account",
			Namespace: "default",
		},
	}
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     "torpedo-role",
		APIGroup: "rbac.authorization.k8s.io",
	}

	envVars := make([]corev1.EnvVar, 0)
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	if len(torpedoSSHUser) != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TORPEDO_SSH_USER",
			Value: torpedoSSHUser,
		})
	}
	if len(torpedoSSHPassword) != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TORPEDO_SSH_PASSWORD",
			Value: torpedoSSHPassword,
		})
	}

	if len(torpedoSSHKey) != 0 {
		file, err := ioutil.ReadFile(torpedoSSHKey)
		if err != err {
			fmt.Printf("failed to read key %s. Cause: %v\n", torpedoSSHKey, err)
			file = []byte("")
		}
		secret := &corev1.Secret{
			Data: map[string][]byte{
				"key4torpedo": []byte(base64.StdEncoding.EncodeToString(file)),
			},
		}
		core.Instance().CreateSecret(secret)
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TORPEDO_SSH_KEY",
			Value: "/home/torpedo/",
		})
		secretVolume := corev1.Volume{
			Name: "ssh-key-volume",
		}
		secretVolumeSource := &corev1.SecretVolumeSource{
			SecretName:  "key4torpedo",
			DefaultMode: &[]int32{256}[0],
		}
		secretVolume.Secret = secretVolumeSource
		volumes = append(volumes, secretVolume)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ssh-key-volume",
			MountPath: "/home/torpedo/",
		})
	}

	if len(azureTenantID) != 0 {
		envVars = append(envVars,
			corev1.EnvVar{
				Name:  "AZURE_TENANT_ID",
				Value: azureTenantID,
			})
	}

	if len(azureClientID) != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "AZURE_CLIENT_ID",
			Value: azureClientID,
		})
	}

	if len(azureClientSecret) != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "AZURE_CLIENT_SECRET",
			Value: azureClientSecret,
		})
	}

	customConfigPath := ""
	if len(customAppConfig) != 0 {
		file, err := ioutil.ReadFile(customAppConfig)
		if err != err {
			fmt.Printf("failed to read file %s. Cause: %v\n", customAppConfig, err)
			file = []byte("")
		}
		configMap := &corev1.ConfigMap{
			Data: map[string]string{
				"custom_app_config.yml": base64.StdEncoding.EncodeToString(file),
			},
		}
		core.Instance().CreateConfigMap(configMap)
		customConfigVolume := corev1.Volume{
			Name: "custom-app-config-volume",
		}
		configMapVolumeSource := &corev1.ConfigMapVolumeSource{}
		configMapVolumeSource.Name = "custom-app-config"
		configMapVolumeSource.Items = []corev1.KeyToPath{{
			Key:  "custom_app_config.yml",
			Path: "custom_app_config.yml",
		}}
		customConfigVolume.ConfigMap = configMapVolumeSource
		volumes = append(volumes, customConfigVolume)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "custom-app-config-volume",
			MountPath: "/mnt/torpedo/custom_app_config.yml",
			SubPath:   "custom_app_config.yml",
		})
		customConfigPath = "/mnt/torpedo/custom_app_config.yml"
	}

	// testresults volume setup
	hostPathVolume := corev1.Volume{
		Name: "testresults",
	}
	hostPathVolumeSource := &corev1.HostPathVolumeSource{
		Path: "/mnt/testresults/",
		Type: &[]corev1.HostPathType{corev1.HostPathDirectoryOrCreate}[0],
	}
	hostPathVolume.HostPath = hostPathVolumeSource
	volumes = append(volumes, hostPathVolume)
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "testresults",
		MountPath: "/testresults/",
	})

	pod := corev1.Pod{}
	pod.Name = "torpedo"
	pod.Labels = map[string]string{
		"app": "torpedo",
	}
	pod.Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Name:            "init-sysctl",
				Image:           "busybox",
				ImagePullPolicy: "IfNotPresent",
				SecurityContext: &corev1.SecurityContext{
					Privileged: &[]bool{true}[0],
				},
				Command: []string{"sh", "-c", "mkdir -p /mnt/testresults && chmod 777 /mnt/testresults/"},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "torpedo",
				Image:           torpedoImg,
				ImagePullPolicy: "Always",
				TTY:             true,
				Command:         []string{"run"},
				Args: []string{
					fmt.Sprintf("--fail-fast=%t", failFast),
					"--skip-tests", skipTests,
					"--focus-tests", focusTests,
					"--test-suite", testSuite,
					"timeout", timeout.String(),
					"--spec-dir", "../drivers/scheduler/k8s/specs",
					"--app-list", appList,
					"--scheduler", scheduler,
					"--log-level", loglevel,
					"--node-driver", nodeDriver,
					"--scale-factor", fmt.Sprintf("%d", scaleFactor),
					"--minimum-runtime-mins", fmt.Sprintf("%d", minRunTime),
					"--driver-start-timeout", driverStartTimeout.String(),
					"--chaos-level", fmt.Sprintf("%d", chaosLevel),
					"--storagenode-recovery-timeout", storagenodeRecoveryTimeout.String(),
					"--provisioner", provisioner,
					"--storage-driver", storageDriver,
					"--config-map", configmap,
					"--custom-config", customConfigPath,
					"--storage-upgrade-endpoint-url", upgradeEndpointUrl,
					"--storage-upgrade-endpoint-version", upgradeEndpointVersion,
					"--destroy-app-timeout", appDestroyTimeout.String(),
				},
				VolumeMounts: volumeMounts,
				Env:          envVars,
			},
		},
		Volumes:            volumes,
		RestartPolicy:      "Never",
		ServiceAccountName: serviceAccount.Name,
		Affinity:           &corev1.Affinity{},
		Tolerations:        []corev1.Toleration{},
	}
	core.Instance().CreateServiceAccount(&serviceAccount)
	rbac.Instance().CreateClusterRole(&clusterRole)
	rbac.Instance().CreateClusterRoleBinding(&clusterRoleBinding)
	core.Instance().CreatePod(&pod)

	if !detach {
		t := func() (interface{}, bool, error) {
			torpedoPod, err := core.Instance().GetPodByName("torpedo", "default")
			if err != nil {
				return nil, true, err
			}
			switch torpedoPod.Status.Phase {
			case corev1.PodRunning:
				return nil, false, nil
			case corev1.PodFailed:
				return nil, false, fmt.Errorf("failed to start torpedo pod. Reason: %s, Message: %s",
					torpedoPod.Status.Reason, torpedoPod.Status.Message)
			default:
				return nil, true, fmt.Errorf("torpedo is not running yet. Status: %s, Reason: %s, Message: %s",
					torpedoPod.Status.Phase, torpedoPod.Status.Reason, torpedoPod.Status.Message)
			}
		}
		_, err := task.DoRetryWithTimeout(t, 10*time.Minute, 5*time.Second)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("torpedo is up and running")
		attachTorpedoStdout()
	}

}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "")
	deployCmd.Flags().IntVarP(&scaleFactor, "scale-factor", "s", 10, "")
	deployCmd.Flags().StringVarP(&scheduler, "scheduler", "S", "k8s", "")
	deployCmd.Flags().StringVarP(&loglevel, "log-level", "L", "debug", "")
	deployCmd.Flags().IntVarP(&chaosLevel, "chaos-level", "C", 5, "")
	deployCmd.Flags().IntVarP(&minRunTime, "minimun-runtime-mins", "m", 0, "")
	deployCmd.Flags().StringVarP(&upgradeEndpointUrl, "storage-upgrade-endpoint-url", "", "", "")
	deployCmd.Flags().StringVarP(&upgradeEndpointVersion, "storage-upgrade-endpoint-version", "", "", "")
	deployCmd.Flags().StringVarP(&provisioner, "provisioner", "P", "", "")
	deployCmd.Flags().StringVarP(&storageDriver, "storage-driver", "D", "", "")
	deployCmd.Flags().StringVarP(&configmap, "config-map", "", "", "")
	deployCmd.Flags().StringVarP(&torpedoImg, "torpedo-img", "", "portworx/torpedo:master-alpha1", "")
	deployCmd.Flags().DurationVarP(&appDestroyTimeout, "destroy-app-timeout", "", 5*time.Minute, "")
	deployCmd.Flags().DurationVarP(&driverStartTimeout, "driver-start-timeout", "", 5*time.Minute, "")
	deployCmd.Flags().DurationVarP(&storagenodeRecoveryTimeout, "storagenode-recovery-timeout", "", 35*time.Minute, "")
	deployCmd.Flags().StringVarP(&azureTenantID, "azure-tenantid", "", "", "")
	deployCmd.Flags().StringVarP(&azureClientID, "azure-clientid", "", "", "")
	deployCmd.Flags().StringVarP(&azureClientSecret, "azure-clientsecret", "", "", "")
	deployCmd.Flags().BoolVarP(&detach, "detach", "d", false, "")
	deployCmd.Flags().StringVarP(&torpedoSSHPassword, "ssh-password", "p", "", "")
	deployCmd.Flags().StringVarP(&torpedoSSHUser, "ssh-user", "l", "", "")
	deployCmd.Flags().StringVarP(&torpedoSSHKey, "ssh-key", "i", "", "")
	deployCmd.Flags().StringVarP(&customAppConfig, "custom-config", "", "", "")
	deployCmd.Flags().StringVarP(&appList, "app-list", "", "", "")
	deployCmd.Flags().StringVarP(&nodeDriver, "node-driver", "", "ssh", "")
	deployCmd.Flags().BoolVarP(&chaosMesh, "enable-chaos", "", false, "")
}
