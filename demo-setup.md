# 🔐 KubeArmor Datasource Setup Guide

This guide outlines how to set up the KubeArmor Grafana datasource plugin, both using **Docker** and within a **Kubernetes** environment.

---

## 🐳 Docker Setup

1. **Start Grafana Container**  
   Run the following command to start Grafana with the KubeArmor plugin mounted:  
   ```bash
   docker compose up
   ```
   This mounts the plugin into `/var/lib/grafana/plugins/` inside the Grafana container.

2. **Access Grafana UI**  
   Open your browser and navigate to:  
   ```
   http://localhost:3000
   ```

3. **Add KubeArmor as a Datasource**  
   - Go to **Configuration → Data Sources**
   - Click **"Add data source"** and choose **KubeArmor**

4. **Configure Backend Credentials**  
   - Enter the required connection details for your backend (OpenSearch or Elasticsearch).
   - Test the connection to ensure Grafana can access your backend.

5. **Explore the Data**  
   - Navigate to **Explore** to query and visualize data from KubeArmor.

---

## ☸️ Kubernetes Setup

1. **Install Grafana in Kubernetes**  
   Follow the [official Helm installation guide](https://grafana.com/docs/grafana/latest/setup-grafana/installation/helm/) to deploy Grafana.

2. **Install the KubeArmor Plugin**  
   - Mount the plugin into the Grafana container using a similar method as in Docker, **or**
   - Download the plugin ZIP from the releases page, extract it, and place it inside the container at:  
     ```
     /var/lib/grafana/plugins
     ```

   > ⚠️ Note: This plugin is **not available via the Grafana plugin catalog**, so manual installation is required.

3. **Complete the Setup**  
   Follow the same steps as the Docker setup to configure the data source and credentials.

4. **Access Grafana**  
   - You may be prompted for login credentials when accessing Grafana. Use the credentials configured during the Helm installation or refer to the setup guide above.

---

## ⚠️ Notes

- Ensure the plugin directory has the **correct permissions** when mounting into the Grafana container.
- Make sure your **backend service (OpenSearch/Elasticsearch)** is running and accessible using the credentials provided.
