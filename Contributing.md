1) Setup a K8s cluster (following vagrant setup for KubeArmor is preferred)
2) Install and setup KubeArmor in the cluster by using this [guide](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) 
3) Clone this repo and build it as mentioned in the README.md 
4) go and edit the grafana.yaml in the grafana-es/deploy directory by specifying the correct plugin path for the 3 path for example 
(path: /home/hari/opensource/grafana_Kubearmor/accuknox-kubearmorplugin-datasource/dist) change this path to correct path to the dist 
for example `path/to/the/plugin/dist` similarly change others as well.
5) Clone the [repo](https://github.com/harisudarsan1/kubearmor-dashboards) and run `kubectl apply -f grafana-es/deploy/`
6) Now open grafana by `kubectl port-forward deploy/grafana -n kubearmor 3000:3000` and open localhost:3000.
7) go to datasources and add `http://elasticsearch:9200` as url and select elasticsearch in the backend field and click test.
8) Now run `kubectl apply -f https://raw.githubusercontent.com/kubearmor/KubeArmor/main/examples/wordpress-mysql/wordpress-mysql-deployment.yaml` to create some deployments monitored 
by Kubermor
9) Do some kubectl exec into these pods and go to grafana create a new dashboard and select the KubeArmor Datasource
you should be able to view the Nodegraph generated from ES logs by Kubermor.
10) Make changes to plugin and make sure plugin working in the grafana and submit a PR with the changes
