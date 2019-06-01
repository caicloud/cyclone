import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Project {
  @observable projectList = null;
  @observable projectDetail = null;
  @observable detailLoading = false;

  @observable resourceList = null;
  @observable loadingResource = false;

  @observable stageList = null;
  @observable loadingStage = false;

  @action.bound
  listProjects(query, cb) {
    fetchApi.listProjects(query).then(data => {
      this.projectList = data;
      cb && cb(data);
    });
  }
  @action.bound
  createProject(data, cb) {
    fetchApi.createProject(data).then(() => {
      cb();
    });
  }

  @action.bound
  updateProject(data, name, cb) {
    fetchApi.updateProject(data, name).then(() => {
      cb();
    });
  }

  @action.bound
  deleteProject(name, cb) {
    fetchApi.removeProject(name).then(() => {
      cb();
    });
  }

  @action.bound
  getProject(name) {
    this.detailLoading = true;
    fetchApi.getProject(name).then(data => {
      this.projectDetail = data;
      this.detailLoading = false;
    });
  }

  @action.bound
  listProjectResources(project) {
    this.loadingResource = true;
    fetchApi.listProjectResources(project).then(data => {
      this.resourceList = data;
      this.loadingResource = false;
    });
  }

  @action.bound
  listProjectStages(project) {
    this.loadingStage = true;
    fetchApi.listProjectStages(project).then(data => {
      this.stageList = data;
      this.loadingStage = false;
    });
  }
}

export default new Project();
