import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Workflow {
  @observable workflowList = {};
  @observable workflowDetail = {};
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
  updateWorkflow(project, workflow, info) {
    return fetchApi.updateWorkflow(project, workflow, info).then(data => {
      this.workflowDetail[workflow] = data;
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
}

export default new Workflow();
