import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Resource {
  @observable resourceList = null;
  @observable resourceDetail = {};
  @observable stageDetail = {};

  @action.bound
  createResource(project, data) {
    fetchApi.createResource(project, data);
  }
  @action.bound
  getResource(project, resource, cb) {
    fetchApi.getResource(project, resource).then(data => {
      this.resourceDetail[`${project}-${resource}`] = data;
      cb && cb(data);
    });
  }

  @action.bound
  createStage(project, data) {
    fetchApi.createStage(project, data);
  }

  @action.bound
  getStage(project, stage, cb) {
    fetchApi.getStage(project, stage).then(data => {
      this.stageDetail[`${project}-${stage}`] = data;
      cb && cb(data);
    });
  }

  @action.bound
  updateStage(project, stage, data) {
    fetchApi.updateStage(project, stage, data);
  }
}

export default new Resource();
