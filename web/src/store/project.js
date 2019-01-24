import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Project {
  @observable projectList = null;
  @observable projectDetail = null;
  @observable detailLoading = false;
  @action.bound
  listProjects() {
    fetchApi.listProjects().then(data => {
      this.projectList = data;
    });
  }
  @action.bound
  createProject(data, cb) {
    fetchApi.createProject(data).then(() => {
      this.listProjects();
      cb();
    });
  }

  @action.bound
  deleteProject(name) {
    fetchApi.removeProject(name).then(() => {
      this.listProjects();
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
}

export default new Project();
