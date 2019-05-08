import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Resource {
  @observable resourceList = null;

  @action.bound
  createResource(project, data) {
    fetchApi.createResource(project, data);
  }

  @action.bound
  createStage(project, data) {
    fetchApi.createStage(project, data);
  }
}

export default new Resource();
