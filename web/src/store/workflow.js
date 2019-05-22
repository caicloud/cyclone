import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Workflow {
  @observable workflowList = {};
  @observable workflowDetail = {};
  @observable workflowRuns = {};

  @action.bound
  listWorklow(projectID) {
    return fetchApi.listWorkflow(projectID, {}).then(data => {
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
  deleteWorkflow(project, workflow) {
    return fetchApi.removeWorkflow(project, workflow).then(() => {
      this.listWorklow(project);
    });
  }
  @action.bound
  getWorkflow(project, workflow) {
    return fetchApi.getWorkflow(project, workflow).then(data => {
      this.workflowDetail[workflow] = data;
    });
  }

  @action.bound
  runWorkflow(project, workflow, info) {
    return fetchApi.runWorkflow(project, workflow, info).then(() => {
      this.listWorkflowRuns(project, workflow);
    });
  }

  @action.bound
  listWorkflowRuns(project, workflow) {
    return fetchApi.listWorkflowRuns(project, workflow).then(data => {
      this.workflowRuns[`${project}-${workflow}`] = data;
    });
  }

  @action.bound
  delelteWorkflowRun(project, workflow, record) {
    return fetchApi.deleteWorkflowRun(project, workflow, record).then(() => {
      this.listWorkflowRuns(project, workflow);
    });
  }
}

export default new Workflow();
