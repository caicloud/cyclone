import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class Workflow {
  @observable workflowList = {};
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
  deleteWorkflow(project, workflow) {
    return fetchApi.removeWorkflow(project, workflow).then(() => {
      this.listWorklow(project);
    });
  }
}

export default new Workflow();
