import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Resource {
  @observable SCMRepos = {};
  @observable resourceTypeList = null;
  @observable resourceTypeDetail = {};
  @observable resourceTypeLoading = false;

  @action.bound
  createStage(project, data, cb) {
    fetchApi.createStage(project, data).then(data => {
      cb && cb(data);
    });
  }

  @action.bound
  getStage(project, stage, cb) {
    fetchApi.getStage(project, stage).then(data => {
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
  deleteStage(project, stage, cb) {
    fetchApi.deleteStage(project, stage).then(() => {
      cb && cb();
    });
  }

  @action.bound
  listSCMRepos(integration) {
    fetchApi.listSCMRepos(integration).then(data => {
      this.SCMRepos[integration] = data;
    });
  }

  @action.bound
  listResourceTypes(operationQuery, cb) {
    fetchApi.listResourceTypes(operationQuery).then(data => {
      this.resourceTypeList = data;
      cb && cb(data);
    });
  }

  @action.bound
  createResourceType(data, cb) {
    fetchApi.createResourceType(data).then(data => {
      cb && cb();
    });
  }

  @action.bound
  getResourceType(resourceType) {
    this.resourceTypeLoading = true;
    fetchApi.getResourceType(resourceType).then(data => {
      this.resourceTypeDetail = data;
      this.resourceTypeLoading = false;
    });
  }

  @action.bound
  updateResourceType(resourceType, data, cb) {
    this.resourceTypeLoading = true;
    fetchApi.updateResourceType(resourceType, data).then(data => {
      this.resourceTypeLoading = false;
      cb && cb();
    });
  }

  @action.bound
  deleteResourceType(resourceType, cb) {
    this.resourceTypeLoading = true;
    fetchApi.deleteResourceType(resourceType).then(data => {
      this.resourceTypeLoading = false;
      cb && cb();
    });
  }
}

export default new Resource();
