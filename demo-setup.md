steps for demo setup KubeArmor Datasource with sample KubeArmor logs
1) Build the plugin as mentioned in the README.md
2) run `docker-compose up` to start grafana and elasticsearch
3) run postlogs.sh to post KubeArmor logs to elasticsearch. Make sure you have `jq` installed for 
parsing
4) Go to datasources section in the url field add `http://elasticsearch:9200` and select Elasticsearch in 
Backend option and click test datasource. If sucessful follow the next steps.
5) Go to dashboard and create visualization with the KubeArmor datasource and select NodeGraph panel to view the graph.

This video illustrates the steps mentioned above, making the process easier to understand.


https://github.com/harisudarsan1/accuknox-kubearmor-datasource/assets/97289088/21b776b7-c809-439b-b723-074cbacf5df2


<video src="assets/grafana-nosound.mp4" width="320" height="240" controls></video>






