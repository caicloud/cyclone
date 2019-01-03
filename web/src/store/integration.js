import { observable, action, keys } from 'mobx';
import fetchApi from '../api/index.js';

class Integration {
  @observable integrationList = [];
  @observable groupIntegrationList = {
    SonarQube: [],
    SCM: [],
    DockerRegistry: [],
  };
  @observable groupKeys = [];
  @observable listLoading = true;
  @action.bound
  getIntegrationList() {
    fetchApi.fetchIntegrationList({}).then(data => {
      this.integrationList = data;
      this.listLoading = false;
      _.forEach(data, o => {
        const type = _.get(o, 'spec.type');
        if (!this.groupIntegrationList[type]) {
          this.groupIntegrationList[type] = [o];
        } else {
          this.groupIntegrationList[type].push(o);
        }
      });
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
}

export default new Integration();
