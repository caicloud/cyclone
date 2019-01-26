import { observable, action, keys } from 'mobx';
import fetchApi from '../api/index.js';

class Integration {
  @observable integrationList = null;
  @observable groupIntegrationList = {
    SonarQube: [],
    SCM: [],
    DockerRegistry: [],
  };
  @observable groupKeys = [];
  @action.bound
  getIntegrationList() {
    fetchApi.fetchIntegrationList({}).then(data => {
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
      this.integrationList = data.items; //TODO 处理返回的集成list items
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
      //this.listProjects();
      cb();
    });
  }
}

export default new Integration();
