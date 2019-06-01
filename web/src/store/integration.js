import { observable, action, keys } from 'mobx';
import fetchApi from '../api/index.js';

class Integration {
  @observable integrationList = null;
  @observable groupIntegrationList = {
    SonarQube: [],
    SCM: [],
    DockerRegistry: [],
    Cluster: [],
  };
  @observable integrationDetail = null;
  @observable detailLoading = false;
  @observable processing = false;

  @action.bound
  getIntegrationList(query = {}) {
    fetchApi.fetchIntegrationList(query).then(data => {
      const groups = {
        SonarQube: [],
        SCM: [],
        DockerRegistry: [],
      };
      _.forEach(data.items, o => {
        const type = _.get(o, 'spec.type');
        if (_.includes(['SonarQube', 'SCM', 'DockerRegistry'], type)) {
          if (!groups[type]) {
            groups[type] = [o];
          } else {
            groups[type].push(o);
          }
        }
      });
      this.integrationList = data.items || [];
      this.groupIntegrationList = groups;
    });
  }

  @action.bound
  getGroupKeys() {
    return keys(this.groupIntegrationList);
  }
  @action.bound
  getGroupItem(key) {
    const keys = key.split('/');
    const items = this.groupIntegrationList[keys[0]];
    return _.find(items, o => _.get(o, 'metadata.name') === keys[1]);
  }
  @action.bound
  createIntegration(data, cb) {
    fetchApi.createIntegration(data).then(() => {
      cb();
    });
  }
  @action.bound
  updateIntegration(data, name, cb) {
    fetchApi.updateIntegration(data, name).then(() => {
      cb();
    });
  }
  @action.bound
  deleteIntegration(name, cb) {
    fetchApi.removeIntegration(name).then(() => {
      cb();
    });
  }
  @action.bound
  getIntegration(name) {
    this.detailLoading = true;
    fetchApi.getIntegration(name).then(data => {
      this.integrationDetail = data;
      this.detailLoading = false;
    });
  }
  @action.bound
  closeCluster(name) {
    this.processing = true;
    fetchApi.closeCluster(name).then(data => {
      this.processing = false;
    });
  }
  @action.bound
  openCluster(name) {
    this.processing = true;
    fetchApi.openCluster(name).then(data => {
      this.processing = false;
    });
  }
  @action.bound
  resetIntegration() {
    this.integrationDetail = null;
  }
}

export default new Integration();
