# KubeArmor Datasource for grafana

[![Slack](https://img.shields.io/badge/Join%20Our%20Community-Slack-blue)](https://join.slack.com/t/kubearmor/shared_invite/zt-2bhlgoxw1-WTLMm_ica8PIhhNBNr2GfA)
[![CI](https://github.com/kubearmor/grafana-datasource/actions/workflows/ci.yml/badge.svg)](https://github.com/kubearmor/grafana-datasource/actions/workflows/ci.yml)
[![Release](https://github.com/kubearmor/grafana-datasource/actions/workflows/release.yml/badge.svg)](https://github.com/kubearmor/grafana-datasource/actions/workflows/release.yml)


This plugin converts KubeArmor logs to grafana Nodegraphs

## Dependencies and Requirements
- Grafana 9+
- Node '>=18 <=20' 
- mage and go
- docker 
- docker-compose
- jq


## Building the Plugin

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/developers/plugin-tools/introduction/grafana-plugin-sdk-for-go) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

### Frontend

1. Install dependencies

   ```bash
   npm install
   ```

2. Build plugin in development mode and run in watch mode

   ```bash
   npm run dev
   ```

3. Build plugin in production mode

   ```bash
   npm run build
   ```

## Plugin Demo
follow the instructions given in the demo-setup.md


## Download

1) Download from the plugin catalog(not ready)
2) Download from [github releases](https://github.com/kubearmor/grafana-datasource/releases)


## Contributing
Follow the instructions provided in the [Contributing.md](Contributing.md) to setup the dev environment.






