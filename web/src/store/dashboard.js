import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Dashboard {
  @observable loading = false;
  @observable storageUsage = {};
  @observable projects = {};
  @observable workflows = {};
  @observable workflowRuns = {};

  @action.bound
  getStorageUsage(cb) {
    this.loading = true;
    fetchApi.getStorageUsage().then(data => {
      this.loading = false;
      this.storageUsage = data;
      cb && cb();
    });
  }

  @action.bound
  getProjects(cb) {
    this.loading = true;
    fetchApi.listProjects().then(data => {
      this.loading = false;
      this.projects = data;
      cb && cb();
    });
  }

  @action.bound
  getWorkflows(cb) {
    this.loading = true;
    fetchApi.listAllWorkflows().then(data => {
      this.loading = false;
      this.workflows = data;
      cb && cb();
    });
  }

  @action.bound
  getWorkflowRuns(cb) {
    this.loading = true;
    fetchApi.listAllWorkflowRuns().then(data => {
      this.loading = false;
      this.workflowRuns = data;
      cb && cb();
    });
  }
}

export default new Dashboard();
