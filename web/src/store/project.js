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

  @observable resourceDetail = {};

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
  listProjectResources(project, cb) {
    this.loadingResource = true;
    fetchApi.listProjectResources(project).then(data => {
      this.resourceList = data;
      this.loadingResource = false;
      cb && cb(data);
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

  @action.bound
  getResource(project, resource, cb) {
    fetchApi.getResource(project, resource).then(data => {
      this.resourceDetail[resource] = data;
      cb && cb(data);
    });
  }

  @action.bound
  updateResource(project, resource, info, cb) {
    fetchApi.updateResource(project, resource, info).then(() => {
      cb && cb();
      this.listProjectResources(project);
    });
  }

  @action.bound
  createResource(project, data, cb) {
    fetchApi.createResource(project, data).then(() => {
      cb && cb();
      this.listProjectResources(project);
    });
  }

  @action.bound
  deleteResource(project, name) {
    fetchApi.deleteResource(project, name).then(() => {
      this.listProjectResources(project);
    });
  }
}

export default new Project();
