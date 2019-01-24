import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Project {
  @observable projectList = null;
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

  deleteProject(name) {
    fetchApi.removeProject(name).then(() => {
      this.listProjects();
    });
  }
}

export default new Project();
