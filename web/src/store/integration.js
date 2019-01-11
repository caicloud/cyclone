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
  @observable listFetched = false;
  @action.bound
  getIntegrationList() {
    fetchApi.fetchIntegrationList({}).then(data => {
      this.listFetched = true;
      const groups = {};
      _.forEach(data, o => {
        const type = _.get(o, 'spec.type');
        if (!this.groupIntegrationList[type]) {
          groups[type] = [o];
        } else {
          groups[type].push(o);
        }
      });
      this.integrationList = data;
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
}

export default new Integration();
