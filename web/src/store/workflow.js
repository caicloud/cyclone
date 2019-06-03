import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Workflow {
  @observable workflowList = {};
  @observable workflowDetail = {};
  @observable workflowRuns = {};

  @action.bound
  listWorklow(projectID, query) {
    return fetchApi.listWorkflow(projectID, query).then(data => {
      this.workflowList[projectID] = data;
    });
  }

  @action.bound
  createWorkflow(project, info) {
    return fetchApi.createWorkflow(project, info);
  }

  @action.bound
  updateWorkflow(project, workflow, info, cb) {
    return fetchApi.updateWorkflow(project, workflow, info).then(data => {
      this.workflowDetail[workflow] = data;
      cb && cb();
    });
  }

  @action.bound
  deleteWorkflow(project, workflow, cb) {
    return fetchApi.removeWorkflow(project, workflow).then(() => {
      this.listWorklow(project);
      cb && cb();
    });
  }
  @action.bound
  getWorkflow(project, workflow) {
    return fetchApi.getWorkflow(project, workflow).then(data => {
      this.workflowDetail[workflow] = data;
    });
  }

  @action.bound
  runWorkflow(project, workflow, info, listQuery = {}) {
    return fetchApi.runWorkflow(project, workflow, info).then(() => {
      this.listWorkflowRuns(project, workflow, listQuery);
    });
  }

  @action.bound
  listWorkflowRuns(project, workflow, query = {}, cb) {
    return fetchApi.listWorkflowRuns(project, workflow, query).then(data => {
      this.workflowRuns[`${project}-${workflow}`] = data;
      cb && cb(data);
    });
  }

  @action.bound
  delelteWorkflowRun(project, workflow, record, listQuery = {}) {
    return fetchApi.deleteWorkflowRun(project, workflow, record).then(() => {
      this.listWorkflowRuns(project, workflow, listQuery);
    });
  }
}

export default new Workflow();
