import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Resource {
  @observable resourceList = null;
  @observable resourceDetail = {};
  @observable stageDetail = {};
  @observable SCMRepos = {};

  @action.bound
  createResource(project, data, cb) {
    fetchApi.createResource(project, data).then(data => {
      cb && cb();
    });
  }
  @action.bound
  getResource(project, resource, cb) {
    fetchApi.getResource(project, resource).then(data => {
      this.resourceDetail[`${project}-${resource}`] = data;
      cb && cb(data);
    });
  }

  @action.bound
  updateResource(project, resource, info) {
    fetchApi.updateResource(project, resource, info);
  }

  @action.bound
  createStage(project, data, cb) {
    fetchApi.createStage(project, data).then(data => {
      cb && cb(data);
    });
  }

  @action.bound
  getStage(project, stage, cb) {
    fetchApi.getStage(project, stage).then(data => {
      this.stageDetail[`${project}-${stage}`] = data;
      cb && cb(data);
    });
  }

  @action.bound
  updateStage(project, stage, data, cb) {
    fetchApi.updateStage(project, stage, data).then(() => {
      cb && cb();
    });
  }

  @action.bound
  listSCMRepos(integration) {
    fetchApi.listSCMRepos(integration).then(data => {
      this.SCMRepos[integration] = data;
    });
  }
}

export default new Resource();
